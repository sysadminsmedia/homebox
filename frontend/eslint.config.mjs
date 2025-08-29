import withNuxt from './.nuxt/eslint.config.mjs';

import tailwind from 'eslint-plugin-tailwindcss';
import prettier from 'eslint-plugin-prettier';

import { includeIgnoreFile } from '@eslint/compat';
import { fileURLToPath } from 'node:url';

const gitignorePath = fileURLToPath(new URL('../.gitignore', import.meta.url));

export default withNuxt([
  includeIgnoreFile(gitignorePath, 'Imported ../.gitignore patterns'),
  ...tailwind.configs['flat/recommended'],
  {
    plugins: {
      prettier,
    },
    rules: {
      'vue/no-undef-components': [
        'error',
        {
          ignorePatterns: [],
        },
      ],
      'no-console': 0,
      'no-unused-vars': 'off',
      'vue/multi-word-component-names': 'off',
      'vue/no-setup-props-destructure': 0,
      'vue/no-multiple-template-root': 0,
      'vue/no-v-model-argument': 0,
      'vue/no-v-html': 0,
      'tailwind/no-custom-classname': 'warn',

      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          ignoreRestSiblings: true,
          destructuredArrayIgnorePattern: '_',
          caughtErrors: 'none',
        },
      ],

      'prettier/prettier': [
        'warn',
        {
          arrowParens: 'avoid',
          semi: true,
          tabWidth: 2,
          useTabs: false,
          vueIndentScriptAndStyle: true,
          singleQuote: false,
          trailingComma: 'es5',
          printWidth: 120,
        },
      ],
    },
  },
]);
