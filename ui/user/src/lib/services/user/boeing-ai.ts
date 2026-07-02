/**
 * Boeing AI credential service - manages per-user UDAL PAT storage.
 *
 * The UDAL PAT (Personal Access Token) is stored per-user and used as the
 * API key for all Boeing AI (BCAI) API calls. Each user must configure their
 * own PAT for audit traceability.
 */

import { doDelete, doGet, doPost } from '../http';

const BOEING_AI_CREDENTIAL_PATH = '/boeing-ai-credential';

export interface BoeingAICredentialStatus {
	configured: boolean;
	issuedAt?: string;
	expiresAt?: string;
	maskedToken?: string;
}

/**
 * Get the current user's Boeing AI credential status.
 * Returns whether a PAT is configured, and metadata about it.
 */
export async function getBoeingAICredentialStatus(opts?: {
	fetch?: typeof fetch;
}): Promise<BoeingAICredentialStatus> {
	try {
		const response = (await doGet(BOEING_AI_CREDENTIAL_PATH, {
			...opts,
			dontLogErrors: true
		})) as BoeingAICredentialStatus;
		return response;
	} catch {
		return { configured: false };
	}
}

/**
 * Save the user's UDAL PAT as their Boeing AI credential.
 * The token is validated against the Boeing Security API before being stored.
 */
export async function saveBoeingAICredential(token: string): Promise<{
	success: boolean;
	error?: string;
}> {
	try {
		await doPost(BOEING_AI_CREDENTIAL_PATH, { token });
		return { success: true };
	} catch (error) {
		const message =
			error instanceof Error ? error.message : 'Failed to save credential';
		return { success: false, error: message };
	}
}

/**
 * Delete the user's stored Boeing AI credential.
 */
export async function deleteBoeingAICredential(): Promise<void> {
	await doDelete(BOEING_AI_CREDENTIAL_PATH);
}
