import {config} from '@dotenvx/dotenvx'

config({
    path: `${__dirname}/../.env`,
})

export const esConfig = {
    awsAccount: process.env.ES_AWS_ACCOUNT,
    awsRegion: process.env.ES_AWS_REGION,
}

console.log('Running with config', JSON.stringify(esConfig, null, 2))
