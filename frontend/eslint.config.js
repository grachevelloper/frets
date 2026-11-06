import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import hooks from 'eslint-plugin-react-hooks';
import refresh from 'eslint-plugin-react-refresh';

export default tseslint.config(
  {ignores: ['dist', 'node_modules']},
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['src/**/*.{ts,tsx}'],
    plugins: {'react-hooks': hooks, 'react-refresh': refresh},
    rules: {
      ...hooks.configs.recommended.rules,
      'react-refresh/only-export-components': ['error', {allowConstantExport: true}],
    },
  },
  {
    files: ['src/shared/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': ['error', {patterns: ['**/modules/**', '**/pages/**', '**/app/**']}],
    },
  },
  {
    files: ['src/modules/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': ['error', {patterns: ['**/pages/**', '**/app/**', '**/modules/**']}],
    },
  },
);
