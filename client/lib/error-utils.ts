/**
 * Safely extracts error message from various error types
 * Handles Axios errors, standard Error objects, and string errors
 */
export function getErrorMessage(error: unknown): string {
  if (!error) return 'An unknown error occurred';
  
  if (typeof error === 'string') {
    return error;
  }
  
  if (error instanceof Error) {
    return error.message;
  }
  
  // Handle Axios error objects
  if (typeof error === 'object' && error !== null) {
    const errorObj = error as any;
    
    // Axios error structure
    if (errorObj.response?.data?.message) {
      return errorObj.response.data.message;
    }
    
    if (errorObj.response?.data?.error) {
      return errorObj.response.data.error;
    }
    
    if (errorObj.message) {
      return errorObj.message;
    }
    
    // Generic object with message property
    if (errorObj.error) {
      return String(errorObj.error);
    }
  }
  
  return 'An unknown error occurred';
}
