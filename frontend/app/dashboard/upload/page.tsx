"use client";

import { useState, useRef } from "react";
import apiClient from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useToast } from "@/components/ui/toast";
import { Upload, File, X } from "lucide-react";

export default function UploadPage() {
    const { addToast } = useToast();
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [uploading, setUploading] = useState(false);
    const [uploadedFiles, setUploadedFiles] = useState<{ url: string; filename: string; size: number }[]>([]);
    const [dragActive, setDragActive] = useState(false);

    const handleUpload = async (file: File) => {
        setUploading(true);
        try {
            const formData = new FormData();
            formData.append("file", file);
            const res = await apiClient.post("/api/v1/upload", formData, {
                headers: { "Content-Type": "multipart/form-data" },
            });
            setUploadedFiles((prev) => [...prev, res.data.data]);
            addToast({ title: "上传成功", description: file.name, variant: "success" });
        } catch (err: unknown) {
            const axiosErr = err as { response?: { data?: { message?: string } } };
            addToast({ title: "上传失败", description: axiosErr.response?.data?.message || "请重试", variant: "error" });
        } finally {
            setUploading(false);
        }
    };

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) handleUpload(file);
    };

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setDragActive(false);
        const file = e.dataTransfer.files?.[0];
        if (file) handleUpload(file);
    };

    const formatSize = (bytes: number) => {
        if (bytes < 1024) return bytes + " B";
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
        return (bytes / (1024 * 1024)).toFixed(1) + " MB";
    };

    return (
        <div className="space-y-6">
            <h2 className="text-2xl font-bold text-gray-900">文件上传</h2>

            <Card>
                <CardContent className="p-6">
                    <div
                        className={`border-2 border-dashed rounded-xl p-12 text-center transition-colors cursor-pointer
              ${dragActive ? "border-blue-500 bg-blue-50" : "border-gray-300 hover:border-gray-400"}`}
                        onClick={() => fileInputRef.current?.click()}
                        onDragOver={(e) => { e.preventDefault(); setDragActive(true); }}
                        onDragLeave={() => setDragActive(false)}
                        onDrop={handleDrop}
                    >
                        <Upload className="h-10 w-10 mx-auto text-gray-400 mb-4" />
                        <p className="text-gray-600 font-medium">点击或拖拽文件到此区域上传</p>
                        <p className="text-sm text-gray-400 mt-2">支持 JPG、PNG、GIF、PDF、DOC 格式，最大 10MB</p>
                        <input ref={fileInputRef} type="file" className="hidden" onChange={handleFileChange} />
                        {uploading && (
                            <div className="mt-4">
                                <div className="animate-spin h-6 w-6 mx-auto border-2 border-blue-600 border-t-transparent rounded-full" />
                            </div>
                        )}
                    </div>
                </CardContent>
            </Card>

            {uploadedFiles.length > 0 && (
                <Card>
                    <CardHeader>
                        <CardTitle>已上传文件</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-2">
                            {uploadedFiles.map((file, i) => (
                                <div key={i} className="flex items-center justify-between rounded-lg border p-3">
                                    <div className="flex items-center gap-3">
                                        <File className="h-5 w-5 text-gray-400" />
                                        <div>
                                            <p className="text-sm font-medium">{file.filename}</p>
                                            <p className="text-xs text-gray-400">{formatSize(file.size)}</p>
                                        </div>
                                    </div>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        onClick={() => setUploadedFiles((prev) => prev.filter((_, idx) => idx !== i))}
                                    >
                                        <X className="h-4 w-4" />
                                    </Button>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            )}
        </div>
    );
}
