import {Argv} from "yargs";

import {getQuerySdk} from "../../../helpers/sdk-helpers";
import {QueryMyKeysRequest} from "@bluzelle/sdk-js/lib/codec/crud/query";

export const command = 'myKeys <address> <uuid>'
export const desc = 'Read all keys in uuid owned by given address'
export const builder = (yargs: Argv) => {
    return yargs
        .help()
}
export const handler = (argv: QueryMyKeysRequest) => {
    return getQuerySdk()
        .then(sdk => sdk.db.q.MyKeys({
            address: argv.address,
            uuid: argv.uuid
        }))
        .then(data => data.keys)
        .then(console.log)
}