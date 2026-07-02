import { BoeingbotService } from '$lib/services';
import type { PageLoad } from './$types';

export const ssr = false;

export const load: PageLoad = async ({ fetch }) => {
	const publishedWorkflows = await BoeingbotService.listPublishedWorkflows({ fetch });
	return { publishedWorkflows };
};
