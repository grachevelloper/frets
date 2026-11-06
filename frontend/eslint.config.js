import {fileURLToPath} from 'node:url';

import js from '@eslint/js';
import {defineConfig} from 'eslint/config';
import i18nJsonPlugin from 'eslint-plugin-i18n-json';
import i18nextPlugin from 'eslint-plugin-i18next';
import importPlugin from 'eslint-plugin-import';
import prettierPlugin from 'eslint-plugin-prettier';
import reactPlugin from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import simpleImportSort from 'eslint-plugin-simple-import-sort';
import globals from 'globals';
import tseslint from 'typescript-eslint';

export default defineConfig([
    {ignores: ['dist/**', 'node_modules/**']},
    {
        files: ['**/*.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'],
        plugins: {
            js,
            '@typescript-eslint': tseslint.plugin,
            'simple-import-sort': simpleImportSort,
            import: importPlugin,
            prettier: prettierPlugin,
            'react-hooks': reactHooks,
            react: reactPlugin,
            i18next: i18nextPlugin,
            'i18n-json': i18nJsonPlugin,
        },
        extends: [js.configs.recommended],
        settings: {
            react: {
                version: 'detect',
            },
            'import/resolver': {typescript: true},
        },
        languageOptions: {
            globals: {...globals.browser, ...globals.node, ...globals.jest},
            parserOptions: {
                ecmaVersion: 'latest',
                sourceType: 'module',
                ecmaFeatures: {
                    jsx: true,
                },
            },
        },
    },
    {...tseslint.configs.recommendedTypeChecked[0], files: ['**/*.{ts,tsx}']},
    ...tseslint.configs.recommendedTypeChecked
        .slice(1)
        .map((config) => ({...config, files: ['**/*.{ts,tsx}']})),
    {
        files: ['**/*.{ts,tsx}'],
        languageOptions: {
            parserOptions: {
                projectService: true,
            },
        },
    },
    {...reactPlugin.configs.flat.recommended, files: ['**/*.{jsx,tsx}']},
    {
        files: ['**/*.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'],
        rules: {
            '@typescript-eslint/no-unsafe-call': 'off',
            'no-case-declarations': 'off',
            '@typescript-eslint/only-throw-error': 'off',
            '@typescript-eslint/no-unused-vars': 'warn',
            'react/display-name': 'off',
            'react/jsx-key': 'off',
            '@typescript-eslint/no-unsafe-assignment': 'off',
            '@typescript-eslint/no-unsafe-member-access': 'off',
            '@typescript-eslint/explicit-function-return-type': 'off',
            '@typescript-eslint/no-explicit-any': 'off',
            '@typescript-eslint/no-floating-promises': 'off',
            '@typescript-eslint/no-unsafe-argument': 'off',
            '@typescript-eslint/no-unsafe-return': 'off',
            'prettier/prettier': 'error',
            'prefer-const': 'error',
            'no-console': 'warn',
            'simple-import-sort/imports': 'off',
            'simple-import-sort/exports': 'off',
            'import/order': [
                'error',
                {
                    groups: [
                        'builtin',
                        'external',
                        'internal',
                        'parent',
                        'sibling',
                        'index',
                        'object',
                    ],
                    pathGroups: [
                        {
                            pattern: '@/**',
                            group: 'internal',
                            position: 'before',
                        },
                    ],
                    pathGroupsExcludedImportTypes: ['internal'],
                    'newlines-between': 'always',
                    alphabetize: {
                        order: 'asc',
                        caseInsensitive: true,
                    },
                },
            ],
        },
    },
    {
        files: ['**/*.{jsx,tsx}'],
        rules: {
            'react/react-in-jsx-scope': 'off',
            'react/prop-types': 'off',
            'react/jsx-uses-react': 'off',
            'i18next/no-literal-string': ['error', {markupOnly: true}],
            'react/jsx-key': 'error',
            'react/jsx-no-duplicate-props': 'error',
            'object-curly-spacing': ['warn', 'never'],
        },
    },
    {
        files: ['**/webpack.config.js', '**/*.config.*', '**/scripts/**'],
        rules: {
            'no-console': 'off',
            '@typescript-eslint/no-var-requires': 'off',
        },
    },
    {
        files: ['**/*.test.{js,jsx,ts,tsx}', '**/__tests__/**'],
        rules: {
            'no-console': 'off',
        },
    },
    {
        files: ['src/shared/i18n/*.json'],
        plugins: {'i18n-json': i18nJsonPlugin},
        processor: {
            meta: {name: '.json'},
            ...i18nJsonPlugin.processors['.json'],
        },
        rules: {
            'i18n-json/valid-message-syntax': [
                'error',
                {
                    syntax: fileURLToPath(
                        new URL(
                            './scripts/validate-i18n-message.cjs',
                            import.meta.url
                        )
                    ),
                },
            ],
            'i18n-json/valid-json': 'error',
        },
    },
    {
        files: ['**/__test__/mocks/**', '**/__test__/utils/**'],
        rules: {
            '@typescript-eslint/no-var-requires': 'off',
        },
    },
]);
