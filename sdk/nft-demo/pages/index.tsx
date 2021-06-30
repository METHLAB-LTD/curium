import Head from 'next/head'
import styles from '../styles/Home.module.css'
import React, {ChangeEvent, useState} from "react";
import {bluzelle} from '@bluzelle/sdk-js'
import {contentType} from "mime-types";

export default function Home() {
    const [state, setState] = useState<string>('ready')

    const onFileSelected = (ev: ChangeEvent<HTMLInputElement>) => {
        setState('uploading')
        ev.target.files?.[0].arrayBuffer()
            .then(data => bluzelle.helpers.nftHelpers.uploadNft("https://client.sentry.testnet.private.bluzelle.com:1317", data as Uint8Array))
            .then(ctx => fetch(`http://nft.bluzelle.com/api/createNft`, {
                method: 'POST',
                body: JSON.stringify({
                    hash: ctx.hash,
                    mime: contentType(ev.target.files?.[0].name || '')
                })
            }).then(resp => resp.json()).then(({id}) => ({...ctx, id})))
            .then(ctx => setState(`done:${ctx.id}`))
    }


    return (
        <div className={styles.container}>
            <Head>
                <title>Create Next App</title>
                <meta name="description" content="Bluzelle NFT storage demo"/>
                <link rel="icon" href="/favicon.ico"/>
            </Head>

            <main className={styles.main}>
                <h2 className={styles.title}>
                    Bluzelle NFT Reliable Storage
                </h2>
                <div style={{padding: 20}}>
                    <input id="image-file" type="file" onInput={onFileSelected}/>
                </div>
                <div style={{padding: 10}}>
                    {state.includes('done:') ? (
                        <>
                            <NodeLink id={state.replace('done:', '')}/>
                        </>
                    ) : (state)}
                </div>
            </main>

        </div>
    )
}

const NodeLink: React.FC<{ port: number, id: string }> = ({id}) => (
    <>
        <div style={{padding: 5}}>
            <a href={`https://client.sentry.testnet.private.bluzelle.com:1317/nft/data/${id}`} target="_blank">
                https://client.sentry.testnet.private.bluzelle.com:1317/nft/data/{id}
            </a>
        </div>
        <div style={{padding: 5}}>
            <a href={`https://a.client.sentry.testnet.private.bluzelle.com:1317/nft/data/${id}`} target="_blank">
                https://a.client.sentry.testnet.private.bluzelle.com:1317/nft/data/{id}
            </a>
        </div>
        <div style={{padding: 5}}>
            <a href={`https://b.client.sentry.testnet.private.bluzelle.com:1317/nft/data/${id}`} target="_blank">
                https://b.client.sentry.testnet.private.bluzelle.com:1317/nft/data/{id}
            </a>
        </div>
    </>
)


