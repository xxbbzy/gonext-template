"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import type { components } from "@/types/api";
import { client } from "@/lib/api-client.gen";
import type { ItemResponse } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Form, FormControl, FormItem, FormLabel } from "@/components/ui/form";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useToast } from "@/components/ui/toast";
import {
  Plus,
  Search,
  Trash2,
  Edit,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";

type Item = ItemResponse;
type ItemFormData = components["schemas"]["CreateItemRequest"];

const initialFormData: ItemFormData = {
  title: "",
  description: "",
  status: "active",
};

function getApiErrorMessage(error: unknown, fallback: string) {
  if (
    error &&
    typeof error === "object" &&
    "message" in error &&
    typeof error.message === "string" &&
    error.message.trim() !== ""
  ) {
    return error.message;
  }

  return fallback;
}

export default function ItemsPage() {
  const tCommon = useTranslations("common");
  const tItems = useTranslations("items");
  const queryClient = useQueryClient();
  const { addToast } = useToast();
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState("");
  const [searchInput, setSearchInput] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const [editItem, setEditItem] = useState<Item | null>(null);
  const [formData, setFormData] = useState<ItemFormData>(initialFormData);

  const { data, isLoading, error } = useQuery({
    queryKey: ["items", page, keyword],
    queryFn: async () => {
      const { data: res, error: apiError } = await client.GET("/api/v1/items", {
        params: {
          query: { page, page_size: 10, keyword: keyword || undefined },
        },
      });
      if (apiError || !res?.data) {
        throw new Error(getApiErrorMessage(apiError, tCommon("error")));
      }

      return res.data as {
        items: ItemResponse[];
        total: number;
        page: number;
        page_size: number;
        total_pages: number;
      };
    },
  });

  const createMutation = useMutation({
    mutationFn: async (data: ItemFormData) => {
      const { data: res, error: apiError } = await client.POST(
        "/api/v1/items",
        {
          body: data,
        }
      );
      if (apiError || !res?.data) {
        throw new Error(getApiErrorMessage(apiError, tItems("createFailed")));
      }
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["items"] });
      setShowCreate(false);
      setFormData(initialFormData);
      addToast({ title: tItems("createSuccess"), variant: "success" });
    },
    onError: () =>
      addToast({ title: tItems("createFailed"), variant: "error" }),
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: number; data: ItemFormData }) => {
      const { data: res, error: apiError } = await client.PUT(
        "/api/v1/items/{id}",
        {
          params: { path: { id } },
          body: data,
        }
      );
      if (apiError || !res?.data) {
        throw new Error(getApiErrorMessage(apiError, tItems("updateFailed")));
      }
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["items"] });
      setEditItem(null);
      setFormData(initialFormData);
      addToast({ title: tItems("updateSuccess"), variant: "success" });
    },
    onError: () =>
      addToast({ title: tItems("updateFailed"), variant: "error" }),
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      const { error: apiError } = await client.DELETE("/api/v1/items/{id}", {
        params: { path: { id } },
      });
      if (apiError) {
        throw new Error(getApiErrorMessage(apiError, tItems("deleteFailed")));
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["items"] });
      addToast({ title: tItems("deleteSuccess"), variant: "success" });
    },
    onError: () =>
      addToast({ title: tItems("deleteFailed"), variant: "error" }),
  });

  const handleSearch = () => {
    setKeyword(searchInput);
    setPage(1);
  };

  const handleEdit = (item: Item) => {
    setEditItem(item);
    setFormData({
      title: item.title ?? "",
      description: item.description ?? "",
      status: item.status === "inactive" ? "inactive" : "active",
    });
    setShowCreate(false);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (editItem?.id) {
      updateMutation.mutate({ id: editItem.id, data: formData });
      return;
    }

    createMutation.mutate(formData);
  };

  const handleCancel = () => {
    setShowCreate(false);
    setEditItem(null);
    setFormData(initialFormData);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">{tItems("title")}</h2>
        <Button
          onClick={() => {
            setShowCreate(true);
            setEditItem(null);
            setFormData(initialFormData);
          }}
        >
          <Plus className="mr-1 h-4 w-4" /> {tCommon("create")}
        </Button>
      </div>

      <Dialog
        open={showCreate || Boolean(editItem)}
        onOpenChange={(open) => !open && handleCancel()}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editItem ? tItems("editTitle") : tItems("createTitle")}
            </DialogTitle>
          </DialogHeader>
          <Form onSubmit={handleSubmit}>
            <FormItem>
              <FormLabel htmlFor="item-title">{tItems("fieldTitle")}</FormLabel>
              <FormControl>
                <Input
                  id="item-title"
                  value={formData.title}
                  onChange={(e) =>
                    setFormData((prev) => ({ ...prev, title: e.target.value }))
                  }
                  required
                />
              </FormControl>
            </FormItem>
            <FormItem>
              <FormLabel htmlFor="item-description">
                {tItems("fieldDescription")}
              </FormLabel>
              <FormControl>
                <Input
                  id="item-description"
                  value={formData.description ?? ""}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      description: e.target.value,
                    }))
                  }
                />
              </FormControl>
            </FormItem>
            <FormItem>
              <FormLabel htmlFor="item-status">
                {tItems("fieldStatus")}
              </FormLabel>
              <FormControl>
                <select
                  id="item-status"
                  value={formData.status}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      status: e.target.value as "active" | "inactive",
                    }))
                  }
                  className="flex h-10 w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 focus:border-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="active">{tItems("statusActive")}</option>
                  <option value="inactive">{tItems("statusInactive")}</option>
                </select>
              </FormControl>
            </FormItem>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={handleCancel}>
                {tCommon("cancel")}
              </Button>
              <Button
                type="submit"
                loading={createMutation.isPending || updateMutation.isPending}
              >
                {editItem ? tCommon("save") : tCommon("create")}
              </Button>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>

      <div className="flex gap-2">
        <Input
          placeholder={tItems("searchPlaceholder")}
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSearch()}
        />
        <Button variant="outline" onClick={handleSearch}>
          <Search className="h-4 w-4" />
        </Button>
      </div>

      <Card>
        <CardContent className="p-0">
          {error ? (
            <div className="p-6 text-sm text-red-600">
              {error instanceof Error ? error.message : tCommon("error")}
            </div>
          ) : null}
          <Table>
            <TableHeader className="bg-gray-50">
              <TableRow className="hover:bg-gray-50">
                <TableHead>ID</TableHead>
                <TableHead>{tItems("fieldTitle")}</TableHead>
                <TableHead>{tItems("fieldStatus")}</TableHead>
                <TableHead>{tItems("createdAt")}</TableHead>
                <TableHead className="text-right">
                  {tItems("actions")}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="py-8 text-center text-gray-400"
                  >
                    {tCommon("loading")}
                  </TableCell>
                </TableRow>
              ) : (data?.items?.length ?? 0) === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="py-8 text-center text-gray-400"
                  >
                    {tCommon("noData")}
                  </TableCell>
                </TableRow>
              ) : (
                data?.items?.map((item) => (
                  <TableRow key={item.id ?? item.title}>
                    <TableCell className="text-gray-600">{item.id}</TableCell>
                    <TableCell className="font-medium">{item.title}</TableCell>
                    <TableCell>
                      <span
                        className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
                          item.status === "active"
                            ? "bg-green-100 text-green-700"
                            : "bg-gray-100 text-gray-600"
                        }`}
                      >
                        {item.status === "active"
                          ? tItems("statusActive")
                          : tItems("statusInactive")}
                      </span>
                    </TableCell>
                    <TableCell className="text-gray-500">
                      {item.created_at
                        ? new Date(item.created_at).toLocaleDateString()
                        : "-"}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleEdit(item)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() =>
                            item.id && deleteMutation.mutate(item.id)
                          }
                        >
                          <Trash2 className="h-4 w-4 text-red-500" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {data && (data.total_pages ?? 0) > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-gray-500">
            {tItems("totalRecords", { total: data.total ?? 0 })}
          </p>
          <div className="flex gap-1">
            <Button
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => setPage((prev) => prev - 1)}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="flex items-center px-3 text-sm">
              {page} / {data.total_pages}
            </span>
            <Button
              variant="outline"
              size="sm"
              disabled={page >= (data.total_pages ?? 0)}
              onClick={() => setPage((prev) => prev + 1)}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
