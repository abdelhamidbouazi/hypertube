export function getErrorMessage(error: unknown): string {
  if (!error) return "An unknown error occurred";

  if (typeof error === "string") {
    return error;
  }

  if (error instanceof Error) {
    return error.message;
  }

  if (typeof error === "object" && error !== null) {
    const errorObj = error as any;

    if (errorObj.response?.data?.message) {
      return errorObj.response.data.message;
    }

    if (errorObj.response?.data?.error) {
      return errorObj.response.data.error;
    }

    if (errorObj.message) {
      return errorObj.message;
    }

    if (errorObj.error) {
      return String(errorObj.error);
    }
  }

  return "An unknown error occurred";
}
