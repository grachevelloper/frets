import {Box, Flex, Text} from '@gravity-ui/uikit';
import type {PropsWithChildren} from 'react';

import {i18n} from '../../shared/i18n';

import styles from './AppLayout.module.css';

export function AppLayout({children}: PropsWithChildren) {
    return (
        <div className={styles.shell}>
            <Flex
                as='header'
                alignItems='center'
                justifyContent='space-between'
                gap='spacing-6'
                className={styles.header}
            >
                <Flex
                    as='a'
                    href='/'
                    direction='column'
                    gap='spacing-1'
                    className={styles.brand}
                    aria-label={i18n('brand-link')}
                >
                    <Text variant='header-1'>{i18n('brand')}</Text>
                    <Text variant='caption-1'>{i18n('studio')}</Text>
                </Flex>
                <Text
                    variant='caption-2'
                    color='secondary'
                    className={styles.tagline}
                >
                    {i18n('tagline')}
                </Text>
            </Flex>
            <Box as='main' className={styles.main}>
                {children}
            </Box>
            <Box as='footer' paddingBlock='spacing-6'>
                <Text variant='caption-2' color='secondary'>
                    {i18n('footer')}
                </Text>
            </Box>
        </div>
    );
}
