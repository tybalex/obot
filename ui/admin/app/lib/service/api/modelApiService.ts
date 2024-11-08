import { Model, ModelManifest, ModelProvider } from "~/lib/model/models";
import { ApiRoutes } from "~/lib/routers/apiRoutes";
import { request } from "~/lib/service/api/primitives";

async function getModels() {
    const { data } = await request<{ items?: Model[] }>({
        url: ApiRoutes.models.getModels().url,
    });

    // Place default models first
    return (
        data.items?.sort((a, b) => (a.default ? -1 : b.default ? 1 : 0)) ?? []
    );
}
getModels.key = () => ({ url: ApiRoutes.models.getModels().path });

async function getModelById(modelId: string) {
    const { data } = await request<Model>({
        url: ApiRoutes.models.getModelById(modelId).url,
    });

    return data;
}
getModelById.key = (modelId?: string) => {
    if (!modelId) return null;

    return {
        url: ApiRoutes.models.getModelById(modelId).path,
        modelId,
    };
};

async function getModelProviders(onlyConfigured = false) {
    const { data } = await request<{ items?: ModelProvider[] }>({
        url: ApiRoutes.toolReferences.base({ type: "modelProvider" }).url,
    });

    if (onlyConfigured) {
        return (
            data.items?.filter(
                (provider) => provider.modelProviderStatus.configured
            ) ?? []
        );
    }

    return data.items ?? [];
}
getModelProviders.key = (onlyConfigured = false) => ({
    url: ApiRoutes.toolReferences.base({ type: "modelProvider" }).path,
    onlyConfigured,
});

async function createModel(manifest: ModelManifest) {
    await new Promise((resolve) => setTimeout(resolve, 1000));
    const { data } = await request<Model>({
        url: ApiRoutes.models.createModel().url,
        method: "POST",
        data: manifest,
    });

    return data;
}

async function updateModel(modelId: string, manifest: ModelManifest) {
    await new Promise((resolve) => setTimeout(resolve, 1000));

    const { data } = await request<Model>({
        url: ApiRoutes.models.updateModel(modelId).url,
        method: "PUT",
        data: manifest,
    });

    return data;
}

async function deleteModel(modelId: string) {
    await request({
        url: ApiRoutes.models.deleteModel(modelId).url,
        method: "DELETE",
    });
}

export const ModelApiService = {
    getModels,
    getModelById,
    getModelProviders,
    createModel,
    updateModel,
    deleteModel,
};
