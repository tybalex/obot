import useSWR from "swr";

import { ForbiddenError } from "~/lib/service/api/apiErrors";
import { BootstrapApiService } from "~/lib/service/api/bootstrapApiService";
import { VersionApiService } from "~/lib/service/api/versionApiService";

export function useAuthStatus() {
	const getBootstrapStatus = useSWR(
		BootstrapApiService.bootstrapStatus.key(),
		BootstrapApiService.bootstrapStatus,
		{ revalidateIfStale: false }
	);

	const bootstrapEnabled =
		!!getBootstrapStatus.data && getBootstrapStatus.data.enabled;

	const getVersion = useSWR(
		VersionApiService.getVersion.key(),
		VersionApiService.getVersion,
		{ revalidateIfStale: false }
	);

	const authEnabled =
		!getVersion.isLoading &&
		(getVersion.data?.authEnabled ||
			getVersion.error instanceof ForbiddenError); // if version throws a 403, obviosuly authentication is enabled

	const isLoading = getBootstrapStatus.isLoading || getVersion.isLoading;

	return { bootstrapEnabled, authEnabled, isLoading };
}
