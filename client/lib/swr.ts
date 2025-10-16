import useSWR from "swr";

import api from "./api";

const fetcher = (url: string) => {
  return api
    .get(url)
    .then((res) => {
      return res.data;
    })
    .catch((err) => {
      console.error("SWR Error:", url, err);
      throw err;
    });
};

export const useApi = (url: string | null) => {
  return useSWR(url, fetcher, {
    revalidateOnFocus: false,
    revalidateOnReconnect: true,
    dedupingInterval: 2000,
  });
};

export default fetcher;
