"use client";

import { useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { uploadFile } from "@/lib/api-client.gen";
import type { UploadResponse } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useToast } from "@/components/ui/toast";
import { Upload, File, X } from "lucide-react";

export default function UploadPage() {
  const tUpload = useTranslations("upload");
  const { addToast } = useToast();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [uploadedFiles, setUploadedFiles] = useState<UploadResponse[]>([]);
  const [dragActive, setDragActive] = useState(false);

  const handleUpload = async (file: File) => {
    setUploading(true);
    try {
      const { data: res, error: apiError } = await uploadFile(file);
      if (apiError || !res?.data) {
        const errMsg =
          (apiError as { message?: string })?.message || tUpload("retry");
        addToast({
          title: tUpload("failed"),
          description: errMsg,
          variant: "error",
        });
        return;
      }
      setUploadedFiles((prev) => [...prev, res.data as UploadResponse]);
      addToast({
        title: tUpload("success"),
        description: file.name,
        variant: "success",
      });
    } catch {
      addToast({
        title: tUpload("failed"),
        description: tUpload("retry"),
        variant: "error",
      });
    } finally {
      setUploading(false);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleUpload(file);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragActive(false);
    const file = e.dataTransfer.files?.[0];
    if (file) {
      handleUpload(file);
    }
  };

  const formatSize = (bytes: number) => {
    if (bytes < 1024) {
      return `${bytes} B`;
    }
    if (bytes < 1024 * 1024) {
      return `${(bytes / 1024).toFixed(1)} KB`;
    }
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">{tUpload("title")}</h2>

      <Card>
        <CardContent className="p-6">
          <div
            className={`cursor-pointer rounded-xl border-2 border-dashed p-12 text-center transition-colors ${
              dragActive
                ? "border-blue-500 bg-blue-50"
                : "border-gray-300 hover:border-gray-400"
            }`}
            onClick={() => fileInputRef.current?.click()}
            onDragOver={(e) => {
              e.preventDefault();
              setDragActive(true);
            }}
            onDragLeave={() => setDragActive(false)}
            onDrop={handleDrop}
          >
            <Upload className="mx-auto mb-4 h-10 w-10 text-gray-400" />
            <p className="font-medium text-gray-600">
              {tUpload("dropzoneTitle")}
            </p>
            <p className="mt-2 text-sm text-gray-400">
              {tUpload("dropzoneHint")}
            </p>
            <input
              ref={fileInputRef}
              type="file"
              className="hidden"
              onChange={handleFileChange}
            />
            {uploading && (
              <div className="mt-4">
                <div className="mx-auto h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {uploadedFiles.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>{tUpload("uploadedFiles")}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {uploadedFiles.map((file, i) => (
                <div
                  key={i}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div className="flex items-center gap-3">
                    <File className="h-5 w-5 text-gray-400" />
                    <div>
                      <p className="text-sm font-medium">{file.filename}</p>
                      <p className="text-xs text-gray-400">
                        {formatSize(file.size ?? 0)}
                      </p>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() =>
                      setUploadedFiles((prev) =>
                        prev.filter((_, idx) => idx !== i)
                      )
                    }
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
