import { EntityMeta } from "~/lib/model/primitives";
import { ThreadManifest } from "~/lib/model/threads";

type ProjectManifest = ThreadManifest & {
	parentID?: Nullish<string>;
	assistantID: string;
	editor: boolean;
	userID: string;
};

export type Project = EntityMeta & ProjectManifest;
