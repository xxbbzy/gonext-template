"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient, { ApiResponse, PagedData } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useToast } from "@/components/ui/toast";
import { Plus, Search, Trash2, Edit, ChevronLeft, ChevronRight } from "lucide-react";

interface Item {
    id: number;
    title: string;
    description: string;
    status: string;
    user_id: number;
    created_at: string;
    updated_at: string;
}

export default function ItemsPage() {
    const queryClient = useQueryClient();
    const { addToast } = useToast();
    const [page, setPage] = useState(1);
    const [keyword, setKeyword] = useState("");
    const [searchInput, setSearchInput] = useState("");
    const [showCreate, setShowCreate] = useState(false);
    const [editItem, setEditItem] = useState<Item | null>(null);
    const [formData, setFormData] = useState({ title: "", description: "", status: "active" });

    const { data, isLoading } = useQuery({
        queryKey: ["items", page, keyword],
        queryFn: async () => {
            const res = await apiClient.get<ApiResponse<PagedData<Item>>>("/api/v1/items", {
                params: { page, page_size: 10, keyword },
            });
            return res.data.data;
        },
    });

    const createMutation = useMutation({
        mutationFn: (data: typeof formData) => apiClient.post("/api/v1/items", data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["items"] });
            setShowCreate(false);
            setFormData({ title: "", description: "", status: "active" });
            addToast({ title: "创建成功", variant: "success" });
        },
        onError: () => addToast({ title: "创建失败", variant: "error" }),
    });

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: number; data: typeof formData }) =>
            apiClient.put(`/api/v1/items/${id}`, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["items"] });
            setEditItem(null);
            setFormData({ title: "", description: "", status: "active" });
            addToast({ title: "更新成功", variant: "success" });
        },
        onError: () => addToast({ title: "更新失败", variant: "error" }),
    });

    const deleteMutation = useMutation({
        mutationFn: (id: number) => apiClient.delete(`/api/v1/items/${id}`),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["items"] });
            addToast({ title: "删除成功", variant: "success" });
        },
        onError: () => addToast({ title: "删除失败", variant: "error" }),
    });

    const handleSearch = () => {
        setKeyword(searchInput);
        setPage(1);
    };

    const handleEdit = (item: Item) => {
        setEditItem(item);
        setFormData({ title: item.title, description: item.description, status: item.status });
        setShowCreate(false);
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (editItem) {
            updateMutation.mutate({ id: editItem.id, data: formData });
        } else {
            createMutation.mutate(formData);
        }
    };

    const handleCancel = () => {
        setShowCreate(false);
        setEditItem(null);
        setFormData({ title: "", description: "", status: "active" });
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h2 className="text-2xl font-bold text-gray-900">项目管理</h2>
                <Button onClick={() => { setShowCreate(true); setEditItem(null); setFormData({ title: "", description: "", status: "active" }); }}>
                    <Plus className="h-4 w-4 mr-1" /> 新建
                </Button>
            </div>

            {/* Create / Edit Form */}
            {(showCreate || editItem) && (
                <Card>
                    <CardHeader>
                        <CardTitle>{editItem ? "编辑项目" : "新建项目"}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div className="space-y-2">
                                <label className="text-sm font-medium">标题</label>
                                <Input value={formData.title} onChange={(e) => setFormData((p) => ({ ...p, title: e.target.value }))} required />
                            </div>
                            <div className="space-y-2">
                                <label className="text-sm font-medium">描述</label>
                                <Input value={formData.description} onChange={(e) => setFormData((p) => ({ ...p, description: e.target.value }))} />
                            </div>
                            <div className="space-y-2">
                                <label className="text-sm font-medium">状态</label>
                                <select
                                    value={formData.status}
                                    onChange={(e) => setFormData((p) => ({ ...p, status: e.target.value }))}
                                    className="flex h-10 w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm"
                                >
                                    <option value="active">Active</option>
                                    <option value="inactive">Inactive</option>
                                </select>
                            </div>
                            <div className="flex gap-2">
                                <Button type="submit" loading={createMutation.isPending || updateMutation.isPending}>
                                    {editItem ? "保存" : "创建"}
                                </Button>
                                <Button type="button" variant="outline" onClick={handleCancel}>取消</Button>
                            </div>
                        </form>
                    </CardContent>
                </Card>
            )}

            {/* Search */}
            <div className="flex gap-2">
                <Input
                    placeholder="搜索项目..."
                    value={searchInput}
                    onChange={(e) => setSearchInput(e.target.value)}
                    onKeyDown={(e) => e.key === "Enter" && handleSearch()}
                />
                <Button variant="outline" onClick={handleSearch}>
                    <Search className="h-4 w-4" />
                </Button>
            </div>

            {/* Table */}
            <Card>
                <CardContent className="p-0">
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead className="border-b bg-gray-50">
                                <tr>
                                    <th className="px-4 py-3 text-left font-medium text-gray-500">ID</th>
                                    <th className="px-4 py-3 text-left font-medium text-gray-500">标题</th>
                                    <th className="px-4 py-3 text-left font-medium text-gray-500">状态</th>
                                    <th className="px-4 py-3 text-left font-medium text-gray-500">创建时间</th>
                                    <th className="px-4 py-3 text-right font-medium text-gray-500">操作</th>
                                </tr>
                            </thead>
                            <tbody>
                                {isLoading ? (
                                    <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-400">加载中...</td></tr>
                                ) : data?.items?.length === 0 ? (
                                    <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-400">暂无数据</td></tr>
                                ) : (
                                    data?.items?.map((item) => (
                                        <tr key={item.id} className="border-b last:border-0 hover:bg-gray-50">
                                            <td className="px-4 py-3 text-gray-600">{item.id}</td>
                                            <td className="px-4 py-3 font-medium">{item.title}</td>
                                            <td className="px-4 py-3">
                                                <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${item.status === "active" ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-600"
                                                    }`}>
                                                    {item.status}
                                                </span>
                                            </td>
                                            <td className="px-4 py-3 text-gray-500">{new Date(item.created_at).toLocaleDateString()}</td>
                                            <td className="px-4 py-3 text-right">
                                                <div className="flex justify-end gap-1">
                                                    <Button variant="ghost" size="icon" onClick={() => handleEdit(item)}>
                                                        <Edit className="h-4 w-4" />
                                                    </Button>
                                                    <Button variant="ghost" size="icon" onClick={() => deleteMutation.mutate(item.id)}>
                                                        <Trash2 className="h-4 w-4 text-red-500" />
                                                    </Button>
                                                </div>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </CardContent>
            </Card>

            {/* Pagination */}
            {data && data.total_pages > 1 && (
                <div className="flex items-center justify-between">
                    <p className="text-sm text-gray-500">共 {data.total} 条记录</p>
                    <div className="flex gap-1">
                        <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                            <ChevronLeft className="h-4 w-4" />
                        </Button>
                        <span className="flex items-center px-3 text-sm">{page} / {data.total_pages}</span>
                        <Button variant="outline" size="sm" disabled={page >= data.total_pages} onClick={() => setPage((p) => p + 1)}>
                            <ChevronRight className="h-4 w-4" />
                        </Button>
                    </div>
                </div>
            )}
        </div>
    );
}
