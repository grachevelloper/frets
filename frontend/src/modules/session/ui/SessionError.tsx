import {Alert, Button, Flex} from '@gravity-ui/uikit';

import {i18n} from '../../../shared/i18n';

import {SessionLayout} from './SessionLayout';

export function SessionError({onRetry}: {onRetry: () => void}) {
    return (
        <SessionLayout>
            <Flex direction='column' gap='spacing-6' alignItems='flex-start'>
                <div role='alert'>
                    <Alert
                        theme='warning'
                        title={i18n('session-error')}
                        message={i18n('unavailable')}
                    />
                </div>
                <Button view='action' size='l' onClick={onRetry}>
                    {i18n('retry')}
                </Button>
            </Flex>
        </SessionLayout>
    );
}
