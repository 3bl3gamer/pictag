module.exports = {
	parserOptions: {
		ecmaVersion: 2019,
		sourceType: 'module',
	},
	env: {
		browser: true,
		es6: true,
	},
	extends: 'eslint:recommended',
	plugins: ['svelte3'],
	overrides: [
		{
			files: ['**/*.svelte'],
			processor: 'svelte3/svelte3',
		},
	],
	rules: {
		'no-console': 'warn',
		'no-unused-vars': ['error', { vars: 'all', args: 'none' }],
	},
	globals: {
		Atomics: 'readonly',
		SharedArrayBuffer: 'readonly',
		process: 'readonly',
	},
}
