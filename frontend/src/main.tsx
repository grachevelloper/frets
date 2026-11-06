import {StrictMode} from 'react';
import {createRoot} from 'react-dom/client';

import '@gravity-ui/uikit/styles/fonts.css';
import '@gravity-ui/uikit/styles/styles.css';
import './app/styles/theme.css';
import './app/styles/global.css';
import {App} from './app/App';

const root = document.getElementById('root');

if (!root) {
    throw new Error('Application root is missing');
}

createRoot(root).render(
    <StrictMode>
        <App />
    </StrictMode>
);
