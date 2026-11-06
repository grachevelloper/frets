import {Flex, Loader, Text} from '@gravity-ui/uikit';

import {i18n} from '../../../shared/i18n';

import {SessionLayout} from './SessionLayout';
import styles from './SessionLoading.module.css';

export function SessionLoading() {
    return (
        <SessionLayout>
            <Flex
                role='status'
                aria-live='polite'
                direction='column'
                alignItems='center'
                justifyContent='center'
                gap='spacing-6'
                className={styles.status}
            >
                <Loader size='m' />
                <Text variant='body-3' color='secondary'>
                    {i18n('checking')}
                </Text>
            </Flex>
        </SessionLayout>
    );
}
