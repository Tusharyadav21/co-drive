'use server';

import { revalidatePath } from 'next/cache';
import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { prisma } from '../prisma';
import { auth } from '../auth';

interface BetterAuthTestContext {
  $context: Promise<{
    test?: {
      login: (options: { userId: string }) => Promise<{
        cookies: Array<{
          name: string;
          value: string;
          path?: string;
          domain?: string;
          httpOnly?: boolean;
          secure?: boolean;
          sameSite?: string;
          expires?: Date | string;
        }>;
      }>;
    };
  }>;
}

export async function devLoginAction() {
  // Security check: Only allow in development or if DEV_LOGIN_ENABLED is explicitly enabled
  const isEnabled = process.env.DEV_LOGIN_ENABLED === 'true';
  const isDev = process.env.NODE_ENV === 'development';

  if (!isDev && !isEnabled) {
    return { error: 'Unauthorized: Direct developer login is only allowed in local development.' };
  }

  try {
    const email = process.env.DEV_LOGIN_EMAIL || 'dev@codrive.io';
    const name = process.env.DEV_LOGIN_NAME || 'CoDrive Developer';

    // Find or create developer user
    let user = await prisma.user.findUnique({
      where: { email },
    });

    if (!user) {
      user = await prisma.user.create({
        data: {
          email,
          name,
          emailVerified: true,
        },
      });
    }

    // Access the Better Auth context
    const ctx = await (auth as unknown as BetterAuthTestContext).$context;
    if (!ctx || !ctx.test) {
      return {
        error:
          'Better Auth testUtils plugin is not initialized. Please ensure it is enabled in development.',
      };
    }

    // Programmatically log in user using testUtils
    const loginResult = await ctx.test.login({
      userId: user.id,
    });

    if (!loginResult || !loginResult.cookies) {
      return { error: 'Failed to generate developer session cookies.' };
    }

    // Set the cookies using Next.js cookies API
    const cookieStore = await cookies();
    for (const cookie of loginResult.cookies) {
      let sameSite: 'lax' | 'strict' | 'none' | undefined = undefined;
      if (cookie.sameSite) {
        const lower = cookie.sameSite.toLowerCase();
        if (lower === 'lax' || lower === 'strict' || lower === 'none') {
          sameSite = lower;
        }
      }

      // Fix UNIX timestamp in seconds vs milliseconds conversion
      let expiresDate: Date | undefined = undefined;
      if (cookie.expires) {
        if (typeof cookie.expires === 'number') {
          // If UNIX timestamp is in seconds, convert to milliseconds
          const isSeconds = cookie.expires < 10000000000;
          expiresDate = new Date(isSeconds ? cookie.expires * 1000 : cookie.expires);
        } else {
          expiresDate = new Date(cookie.expires);
        }
      }

      // Standard cookie configuration for local/production
      const cookieOptions: {
        path?: string;
        domain?: string;
        httpOnly?: boolean;
        secure?: boolean;
        sameSite?: 'lax' | 'strict' | 'none';
        expires?: Date;
      } = {
        path: cookie.path,
        httpOnly: cookie.httpOnly,
        secure: cookie.secure,
        sameSite: sameSite,
        expires: expiresDate,
      };

      // Omit the "localhost" domain to prevent browsers from rejecting the cookie on local HTTP
      if (cookie.domain && cookie.domain !== 'localhost') {
        cookieOptions.domain = cookie.domain;
      }

      cookieStore.set(cookie.name, cookie.value, cookieOptions);
    }
  } catch (err) {
    console.error('Developer login failed:', err);
    return { error: err instanceof Error ? err.message : 'Developer login failed unexpectedly.' };
  }

  // Redirect on server-side outside of try-catch block to ensure Next.js redirect mechanism executes cleanly
  revalidatePath('/');
  redirect('/');
}
