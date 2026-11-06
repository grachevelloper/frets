import {ApiError, request} from '../../../shared/api/client';
import type {Profile} from '../model/types';

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
