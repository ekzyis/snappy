import { fileURLToPath, pathToFileURL } from 'node:url'
import { dirname, resolve as resolvePath } from 'node:path'

const scriptsDir = dirname(fileURLToPath(import.meta.url))
const snRoot = resolvePath(scriptsDir, '..', 'stacker.news')
const scriptsParent = pathToFileURL(resolvePath(scriptsDir, 'noop.js')).href

function isBare (specifier) {
  return !specifier.startsWith('.') && !specifier.startsWith('/') &&
    !specifier.startsWith('node:') && !specifier.includes('://')
}

export async function resolve (specifier, context, next) {
  if (specifier.startsWith('@/')) {
    let target = resolvePath(snRoot, specifier.slice(2))
    if (!/\.[cm]?jsx?$/.test(target)) target += '.js'
    return { url: pathToFileURL(target).href, shortCircuit: true }
  }
  // The submodule has no node_modules; resolve its bare deps from scripts/.
  if (isBare(specifier)) {
    return next(specifier, { ...context, parentURL: scriptsParent })
  }
  // The submodule's relative imports omit the .js extension.
  if (specifier.startsWith('.') && !/\.[cm]?jsx?$/.test(specifier)) {
    return next(specifier + '.js', context)
  }
  return next(specifier, context)
}
