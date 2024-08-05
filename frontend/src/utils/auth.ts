// src/utils/auth.ts
export const checkAuth = async (): Promise<boolean> => {
  const token = localStorage.getItem('token');
  if (!token) return false;

  try {
    const response = await fetch('http://localhost:3008/check-auth', {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });

    return response.ok;
  } catch (error) {
    console.error('Error checking authentication:', error);
    return false;
  }
};
