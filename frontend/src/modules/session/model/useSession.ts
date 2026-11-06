import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query';
import {getProfile, logout} from '../api/session';

const sessionKey = ['session'];

export function useSession() {
  return useQuery({
    queryKey: sessionKey,
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
      client.removeQueries({predicate: (query) => query.queryKey[0] !== sessionKey[0]});
      client.setQueryData(sessionKey, null);
    },
  });
}
