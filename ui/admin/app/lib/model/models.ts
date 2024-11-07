import { z } from "zod";

import { EntityMeta } from "~/lib/model/primitives";

export type ModelManifest = {
    name?: string;
    targetModel?: string;
    modelProvider: string;
    active: boolean;
    default: boolean;
};

export type ModelProviderStatus = {
    configured: boolean;
    missingEnvVars?: string[];
};

export type Model = EntityMeta & ModelManifest & ModelProviderStatus;

export const ModelManifestSchema = z.object({
    name: z.string(),
    targetModel: z.string().min(1, "Required"),
    modelProvider: z.string().min(1, "Required"),
    active: z.boolean(),
    default: z.boolean(),
});

export type ModelProvider = EntityMeta & {
    description?: string;
    builtin: boolean;
    active: boolean;
    modelProviderStatus: ModelProviderStatus;
    name: string;
    reference: string;
    toolType: "modelProvider";
};
