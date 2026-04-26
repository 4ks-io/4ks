import { auth0 } from '@/libs/auth0';
import { ApiClient } from '@4ks/api-fetch';

export const apiURL = `${process.env.IO_4KS_API_URL}`;

export async function getAPIClient(): Promise<ApiClient> {
  try {
    // authenticated
    const { token } = await auth0.getAccessToken();
    return new ApiClient({
      BASE: apiURL,
      TOKEN: token,
    });
  } catch (e) {
    // anonymous
    return new ApiClient({
      BASE: apiURL,
    });
  }
}
