import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'
import { writeFileSync, mkdirSync } from 'node:fs'
import { printSchema } from 'graphql'
import { makeExecutableSchema } from '@graphql-tools/schema'
import typeDefs from '../stacker.news/api/typeDefs/index.js'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const outDir = resolve(root, 'client', 'testdata')
const outFile = resolve(outDir, 'schema.graphql')

const schema = makeExecutableSchema({ typeDefs })

mkdirSync(outDir, { recursive: true })
writeFileSync(outFile, printSchema(schema) + '\n')

console.log(`wrote ${outFile}`)
