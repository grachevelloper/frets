export class ApiError extends Error {
    constructor(public readonly status: number) {
        super(`Request failed: ${status}`);
    }
}

export async function request(path: string, options?: RequestInit) {
    const response = await fetch(`/api/v1${path}`, {
        ...options,
        credentials: 'same-origin',
    });

    if (!response.ok) {
        throw new ApiError(response.status);
    }

    return response;
}
