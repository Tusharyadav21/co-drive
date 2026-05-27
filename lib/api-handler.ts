import { NextResponse } from 'next/server';
import { ZodError } from 'zod';

/**
 * Error type for known, client-safe failures. The `message` is returned to the
 * caller verbatim with the given HTTP status. Anything thrown that is NOT an
 * ApiError (or ZodError) is treated as an unexpected internal error and the
 * details are kept server-side.
 */
export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

export type ApiHandler<T> = () => Promise<NextResponse | T>;

export async function handleApiRequest<T>(
  requestHandler: ApiHandler<T>,
  errorMessage: string = 'Internal Server Error'
) {
  try {
    const result = await requestHandler();
    if (result instanceof NextResponse) return result;
    return NextResponse.json(result);
  } catch (error) {
    if (error instanceof ApiError) {
      return NextResponse.json({ error: error.message }, { status: error.status });
    }

    if (error instanceof ZodError) {
      return NextResponse.json(
        { error: error.issues[0]?.message ?? 'Invalid request' },
        { status: 400 }
      );
    }

    // Unexpected error: log the full detail server-side, but never leak internal
    // information (DB errors, stack traces, connection strings) to the client.
    console.error(`[API Error] ${errorMessage}:`, error);
    return NextResponse.json({ error: errorMessage }, { status: 500 });
  }
}
