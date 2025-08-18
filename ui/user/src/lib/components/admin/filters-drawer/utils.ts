import type { OrgUser } from '$lib/services';

/**
 * Generates a display name for a user with fallbacks and contextual information.
 *
 * @param users - Map of user IDs to user objects
 * @param id - The ID of the user to get the display name for
 * @param hasConflict - Optional callback function that returns true if there's a naming conflict
 * @returns A formatted display name string
 *
 */
export function getUserDisplayName(
	users: Map<string, OrgUser>,
	id: string,
	hasConflict?: () => boolean
): string {
	const user = users.get(id);

	// Create an array of potential primary display values in order of preference
	const primaryValues = [
		user?.displayName,
		user?.originalUsername,
		user?.originalEmail,
		user?.username,
		user?.email,
		'Unknown User'
	].filter(Boolean);

	let display = primaryValues[0] ?? '';

	// If a conflict detection function is provided and it returns true,
	// add secondary identifier to disambiguate the user
	if (hasConflict?.()) {
		const secondaryValues = [
			user?.email,
			user?.originalEmail,
			user?.username,
			user?.originalUsername
		].filter(Boolean);

		// Find the first secondary value that's available and different from the primary display
		const secondary = secondaryValues.find((name) => !!name && name !== display);

		if (secondary) {
			display = [display, `(${secondary})`].filter(Boolean).join(' ');
		}
	}

	// If the user has been deleted, append a deletion indicator
	if (user?.deletedAt) {
		display += ' (Deleted)';
	}

	return display;
}
