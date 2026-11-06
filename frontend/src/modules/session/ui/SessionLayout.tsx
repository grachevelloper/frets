import {Flex, Text} from '@gravity-ui/uikit';
import type {PropsWithChildren} from 'react';

import {i18n} from '../../../shared/i18n';

import styles from './SessionLayout.module.css';

export function SessionLayout({children}: PropsWithChildren) {
    return (
        <div className={styles.layout}>
            <Flex
                as='section'
                direction='column'
                justifyContent='center'
                gap='spacing-6'
                className={styles.hero}
                aria-labelledby='studio-title'
            >
                <Text
                    as='h1'
                    variant='display-4'
                    id='studio-title'
                    className={styles.title}
                >
                    {i18n('own-rhythm')}
                    <br />
                    {i18n('own-frets')}
                </Text>
                <Text as='p' variant='body-3' className={styles.description}>
                    {i18n('intro')}
                </Text>
                <div className={styles.strings} aria-hidden='true'>
                    <i />
                    <i />
                    <i />
                    <i />
                    <i />
                    <i />
                </div>
            </Flex>
            <Flex
                as='section'
                alignItems='center'
                justifyContent='center'
                className={styles.entry}
            >
                <div className={styles.content}>{children}</div>
            </Flex>
        </div>
    );
}
