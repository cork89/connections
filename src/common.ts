export function isErrorWithProperty(error: unknown, property: string): error is Error & Record<string, unknown> {
    return typeof error === 'object' && error !== null && property in error;
}