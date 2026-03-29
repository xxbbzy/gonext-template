import { mutationOptions, queryOptions } from "@tanstack/react-query";

import {
  createItem,
  deleteItem,
  listItems,
  updateItem,
  type CreateItemRequest,
  type ListItemsQuery,
  type UpdateItemRequest,
} from "@/lib/api-client.gen";

type UpdateItemVariables = {
  id: number;
  body: UpdateItemRequest;
};

function normalizeItemsQuery(
  query: Partial<ListItemsQuery> = {}
): ListItemsQuery {
  return {
    page: query.page ?? 1,
    page_size: query.page_size ?? 10,
    ...(query.keyword ? { keyword: query.keyword } : {}),
    ...(query.status ? { status: query.status } : {}),
  };
}

export const itemKeys = {
  all: ["items"] as const,
  lists: () => [...itemKeys.all, "list"] as const,
  list: (query: Partial<ListItemsQuery> = {}) =>
    [...itemKeys.lists(), normalizeItemsQuery(query)] as const,
};

export function itemsQueryOptions(query: Partial<ListItemsQuery> = {}) {
  const normalizedQuery = normalizeItemsQuery(query);

  return queryOptions({
    queryKey: itemKeys.list(normalizedQuery),
    queryFn: () => listItems(normalizedQuery),
  });
}

export function createItemMutationOptions() {
  return mutationOptions({
    mutationKey: [...itemKeys.all, "create"] as const,
    mutationFn: (body: CreateItemRequest) => createItem(body),
  });
}

export function updateItemMutationOptions() {
  return mutationOptions({
    mutationKey: [...itemKeys.all, "update"] as const,
    mutationFn: ({ id, body }: UpdateItemVariables) => updateItem(id, body),
  });
}

export function deleteItemMutationOptions() {
  return mutationOptions({
    mutationKey: [...itemKeys.all, "delete"] as const,
    mutationFn: (id: number) => deleteItem(id),
  });
}
