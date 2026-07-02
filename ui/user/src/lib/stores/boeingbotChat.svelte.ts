import type { ChatAPI, ChatSession } from '$lib/services/boeingbot/chat/index.svelte';
import type { Chat, Resource } from '$lib/services/boeingbot/types';
import { writable } from 'svelte/store';

export interface BoeingbotChat {
	projectId: string;
	sessionId?: string;
	chat?: ChatSession;
	api: ChatAPI;
	sessions: Chat[];
	isThreadsLoading: boolean;
	resources: Resource[];
}

/**
 * Storing boeingbot chat data in a store so it can be accessed from anywhere in the app.
 */
export const boeingbotChat = writable<BoeingbotChat | null>(null);
