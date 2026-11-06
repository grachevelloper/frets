import {Alert, Button, Flex} from '@gravity-ui/uikit';

import {i18n} from '../../../shared/i18n';
import {useLogout} from '../model/useSession';

export function SignOutButton() {
    const logout = useLogout();

    return (
        <Flex direction='column' gap='spacing-3'>
            {logout.isError && (
                <div role='alert'>
                    <Alert
                        theme='danger'
                        title={i18n('logout-error')}
                        message={i18n('retry-short')}
                    />
                </div>
            )}
            <Button
                view='outlined'
                size='l'
                loading={logout.isPending}
                onClick={() => logout.mutate()}
            >
                {i18n('logout')}
            </Button>
        </Flex>
    );
}
