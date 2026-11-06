import {Button, Flex, Text} from '@gravity-ui/uikit';

import {i18n} from '../../../shared/i18n';

export default function NotFoundPage() {
    return (
        <Flex
            as='section'
            direction='column'
            alignItems='flex-start'
            gap='spacing-6'
            padding='spacing-8'
        >
            <Text as='h1' variant='header-1'>
                {i18n('not-found')}
            </Text>
            <Button href='/' view='outlined'>
                {i18n('home')}
            </Button>
        </Flex>
    );
}
