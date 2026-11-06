import {Flex, Text} from '@gravity-ui/uikit';

import {i18n} from '../../../shared/i18n';
import {SessionLayout} from '../ui/SessionLayout';
import {SignInButtons} from '../ui/SignInButtons';

import styles from './LoginPage.module.css';

export default function LoginPage({returnTo}: {returnTo: string}) {
    return (
        <SessionLayout>
            <Flex direction='column' gap='spacing-6'>
                <Flex direction='column' gap='spacing-3'>
                    <Text
                        as='h2'
                        variant='display-1'
                        id='login-title'
                        className={styles.heading}
                    >
                        {i18n('sign-in-title')}
                    </Text>
                    <Text
                        as='p'
                        variant='body-3'
                        color='secondary'
                        className={styles.copy}
                    >
                        {i18n('choose-account')}
                    </Text>
                </Flex>
                <SignInButtons returnTo={returnTo} />
                <Text
                    as='p'
                    variant='caption-2'
                    color='secondary'
                    className={styles.hint}
                >
                    {i18n('account-hint')}
                </Text>
            </Flex>
        </SessionLayout>
    );
}
