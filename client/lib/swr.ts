import useSWR from "swr";
import { addToast } from "@heroui/toast";
import { getErrorMessage } from "./error-utils";

import api from "./api";

const fetcher = (url: string) => {
  return api
    .get(url)
    .then((res) => {
      return res.data;
    })
    .catch((err) => {
      if (err?.response?.status !== 401) {
        addToast({
          title: "Failed to load data",
          description: getErrorMessage(err),
          severity: "danger",
          timeout: 4000,
        });
      }
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
