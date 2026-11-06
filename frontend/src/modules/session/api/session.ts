import {ApiError, request} from '../../../shared/api/client';
import type {LoginProvider, Profile} from '../model/types';
import {getReturnPath} from '../model/returnPath';

export async function getProfile(signal: AbortSignal): Promise<Profile | null> {
  try {
    const response = await request('/me', {signal});

    return await response.json();
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return null;
    }

    throw error;
  }
}

export async function logout() {
  try {
    await request('/auth/logout', {method: 'POST'});
  } catch (error) {
    if (!(error instanceof ApiError && error.status === 401)) {
      throw error;
    }
  }
}

export function getLoginHref(provider: LoginProvider, path: string) {
  const query = new URLSearchParams({next: getReturnPath(path)});

  return `/api/v1/auth/login/${provider}?${query}`;
}
