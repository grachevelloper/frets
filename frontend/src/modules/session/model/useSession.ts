import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query';

import {getProfile, logout} from '../api/session';

const SESSION_QUERY_KEY = ['session'];

export function useSession() {
    return useQuery({
        queryKey: SESSION_QUERY_KEY,
        queryFn: ({signal}) => getProfile(signal),
        staleTime: 30_000,
    });
}

export function useLogout() {
    const client = useQueryClient();

    return useMutation({
        mutationFn: logout,
        onSuccess: async () => {
            await client.cancelQueries();
            client.removeQueries({
                predicate: (query) =>
                    query.queryKey[0] !== SESSION_QUERY_KEY[0],
            });
            client.setQueryData(SESSION_QUERY_KEY, null);
        },
    });
}
