import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Helmet } from 'react-helmet-async';
import { logger } from '../../api/axiosConfig';

// Get app name from environment variable with fallback
const APP_NAME = process.env.REACT_APP_NAME || 'RecoNya';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

/**
 * ErrorBoundary component to catch JavaScript errors anywhere in the child component tree,
 * log those errors, and display a fallback UI instead of the component tree that crashed.
 */
class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
    errorInfo: null
  };

  static getDerivedStateFromError(error: Error): State {
    // Update state so the next render will show the fallback UI
    return { hasError: true, error, errorInfo: null };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    // Log the error to the console and potentially to a reporting service
    logger.error('Error caught by ErrorBoundary:', {
      error,
      componentStack: errorInfo.componentStack
    });

    this.setState({
      error,
      errorInfo
    });
  }

  render(): ReactNode {
    if (this.state.hasError) {
      // Render custom fallback UI if provided
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // Default fallback UI
      return (
        <div className="error-boundary p-4 text-center">
          <Helmet>
            <title>Something Went Wrong | {APP_NAME}</title>
          </Helmet>
          <div className="card bg-dark text-success border-danger">
            <div className="card-header border-danger">
              <h4>Something went wrong</h4>
            </div>
            <div className="card-body">
              <p>The application encountered an error. Please try refreshing the page.</p>
              {process.env.NODE_ENV === 'development' && this.state.error && (
                <div className="error-details mt-3 text-start">
                  <h5>Error Details:</h5>
                  <pre className="border border-danger p-3 rounded">
                    {this.state.error.toString()}
                  </pre>
                  {this.state.errorInfo && (
                    <div className="mt-3">
                      <h5>Component Stack:</h5>
                      <pre className="border border-danger p-3 rounded">
                        {this.state.errorInfo.componentStack}
                      </pre>
                    </div>
                  )}
                </div>
              )}
              <button
                className="btn btn-success mt-3"
                onClick={() => window.location.reload()}
              >
                Refresh Page
              </button>
            </div>
          </div>
        </div>
      );
    }

    // If there's no error, render children normally
    return this.props.children;
  }
}

export default ErrorBoundary;