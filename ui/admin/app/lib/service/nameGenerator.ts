import { faker } from "@faker-js/faker";

export function generateRandomName(): string {
    // faker doesn't have a "capitalized word" function, so we need to do it manually :(
    const rawAdjective = faker.word.adjective();
    const rawAnimal = faker.animal.type();
    const adjective =
        rawAdjective.charAt(0).toUpperCase() + rawAdjective.slice(1);
    const animal = rawAnimal.charAt(0).toUpperCase() + rawAnimal.slice(1);

    return `${adjective} ${animal}`;
}
