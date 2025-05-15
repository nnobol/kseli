import { writable } from 'svelte/store';
import { getItemFromSessionStorage } from '$lib/api/utils';

let cachedCryptoKey: CryptoKey | null = null;
export const keyStore = writable<string | null>(null);

export function clearCryptoKey() {
    cachedCryptoKey = null;
}

export async function generateEncryptionKey(): Promise<string> {
    const key = await crypto.subtle.generateKey(
        { name: 'AES-GCM', length: 256 },
        true,
        ['encrypt', 'decrypt']
    );
    const raw = await crypto.subtle.exportKey('raw', key);
    return btoa(String.fromCharCode(...new Uint8Array(raw)));
}

export async function importKeyFromString(base64key: string): Promise<CryptoKey> {
    const binary = Uint8Array.from(atob(base64key), c => c.charCodeAt(0));
    return crypto.subtle.importKey(
        'raw',
        binary.buffer,
        'AES-GCM',
        true,
        ['encrypt', 'decrypt']
    );
}

export async function getCryptoKey(): Promise<CryptoKey> {
    if (cachedCryptoKey) return cachedCryptoKey;

    const keyString = getItemFromSessionStorage("encryptionKey");
    if (!keyString) {
        throw new Error("Missing encryption key in session storage.");
    }

    try {
        cachedCryptoKey = await importKeyFromString(keyString);
        return cachedCryptoKey;
    } catch (err) {
        throw new Error("Failed to import encryption key. It may be invalid or corrupted.");
    }
}

export async function encodeMessage(plainText: string): Promise<string> {
    const cryptoKey = await getCryptoKey();

    const iv = crypto.getRandomValues(new Uint8Array(12));
    const encodedText = new TextEncoder().encode(plainText);

    const encrypted = await crypto.subtle.encrypt(
        { name: "AES-GCM", iv },
        cryptoKey,
        encodedText
    );

    const ivString = btoa(String.fromCharCode(...iv));
    const dataString = btoa(String.fromCharCode(...new Uint8Array(encrypted)));

    return `${ivString}:${dataString}`;
}

export async function decodeMessage(packed: string): Promise<string> {
    const cryptoKey = await getCryptoKey();

    const [ivBase64, dataBase64] = packed.split(":");

    if (!ivBase64 || !dataBase64) {
        throw new Error("Invalid message format.");
    }

    const iv = Uint8Array.from(atob(ivBase64), c => c.charCodeAt(0));
    const encryptedData = Uint8Array.from(atob(dataBase64), c => c.charCodeAt(0));


    const decryptedBuffer = await crypto.subtle.decrypt(
        { name: "AES-GCM", iv },
        cryptoKey,
        encryptedData
    );

    return new TextDecoder().decode(decryptedBuffer);
}