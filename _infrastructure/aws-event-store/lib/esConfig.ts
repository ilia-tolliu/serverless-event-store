import {config} from '@dotenvx/dotenvx'

config({
    path: `${__dirname}/../.env`,
})

export const esConfig = {
    awsAccount: process.env.ES_AWS_ACCOUNT,
    awsRegion: process.env.ES_AWS_REGION,
    appMode: process.env.ES_APP_MODE?.toLowerCase() ?? '',
}

if (!['development', 'staging', 'production'].includes(esConfig.appMode)) {
    throw new Error(`Invalid ES_APP_MODE [${esConfig.appMode}]. Should be one of 'development', 'staging', 'production']`)
}

console.log('Running with config', JSON.stringify(esConfig, null, 2))
