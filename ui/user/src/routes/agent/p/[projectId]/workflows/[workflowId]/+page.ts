import { BoeingbotService } from '$lib/services';
import type { PageLoad } from './$types';

export const ssr = false;

export const load: PageLoad = async ({ fetch, params }) => {
	const workflowId = params.workflowId;
	const projectId = params.projectId;
	const publishedWorkflows = await BoeingbotService.listPublishedWorkflows({ fetch });

	return {
		workflowId,
		projectId,
		publishedWorkflows
	};
};
