import svelte from 'rollup-plugin-svelte';
import nodeResolve from 'rollup-plugin-node-resolve'

function mustImport(name) {
	return import(name).catch(err => {
		throw err
	})
}

export default function(commandOptions) {
	const isProd = process.env.NODE_ENV === 'production'

	let devPlugins = []
	if (!isProd)
		devPlugins.push(
			mustImport('rollup-plugin-serve').then(({ default: serve }) =>
				serve({
					contentBase: 'dist',
					host: commandOptions.configHost || 'localhost',
					port: commandOptions.configPort || '12345',
				}),
			),
			mustImport('rollup-plugin-livereload').then(({ default: livereload }) =>
				livereload({ verbose: true, watch: 'dist/bundle.js' }),
			),
		)

	return Promise.all(devPlugins).then(devPlugins => ({
		input: __dirname + '/src/index.js',
		output: {
			format: 'esm',
			dir: 'dist',
			entryFileNames: isProd ? 'bundle.[hash].js' : 'bundle.js',
			sourcemap: true,
		},
		plugins: [
			...devPlugins,
			svelte({
				dev: !isProd,
				css: css => {
					css.write('dist/bundle.css')
				},
			}),
			// commonjs({}), //rollup-plugin-commonjs
			nodeResolve({
				mainFields: (isProd ? [] : ['source']).concat(['module', 'main']),
				dedupe: importee => importee === 'svelte' || importee.startsWith('svelte/'),
			}),
		],
		watch: { clearScreen: false },
	}))
}
