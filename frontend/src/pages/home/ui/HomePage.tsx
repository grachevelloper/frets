import {Flex, Text} from '@gravity-ui/uikit';

import {SignOutButton} from '../../../modules/session';
import type {Profile} from '../../../modules/session';
import {i18n} from '../../../shared/i18n';

export default function HomePage({profile}: {profile: Profile}) {
    return (
        <Flex
            as='section'
            direction='column'
            alignItems='flex-start'
            gap='spacing-6'
            padding='spacing-8'
        >
            <Text variant='caption-2' color='secondary'>
                {i18n('library')}
            </Text>
            <Text as='h1' variant='header-1'>
                {i18n('greeting', {nickname: profile.nickname})}
            </Text>
            <Text as='p' variant='body-2' color='secondary'>
                {i18n('welcome')}
            </Text>
            <SignOutButton />
        </Flex>
    );
}
