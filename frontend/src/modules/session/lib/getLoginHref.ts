import {getReturnPath} from '../model/returnPath';
import type {LoginProvider} from '../model/types';

export function getLoginHref(provider: LoginProvider, path: string) {
    const query = new URLSearchParams({next: getReturnPath(path)});

    return `/api/v1/auth/login/${provider}?${query}`;
}
