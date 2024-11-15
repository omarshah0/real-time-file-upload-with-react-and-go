import React, { useState, useEffect, useRef } from 'react'
import { FaUpload } from 'react-icons/fa'

const CHUNK_SIZE = 64 * 1024 // 64KB chunks

const FileUploader: React.FC = () => {
    const [progress, setProgress] = useState(0)
    const [status, setStatus] = useState('')
    const ws = useRef<WebSocket | null>(null)
    const fileRef = useRef<File | null>(null)
    const offsetRef = useRef(0)

    useEffect(() => {
        ws.current = new WebSocket('ws://localhost:8080/ws')

        ws.current.onopen = () => {
            setStatus('Connected to server')
        }

        ws.current.onmessage = (event) => {
            const message = JSON.parse(event.data)
            if (message.type === 'progress') {
                const currentProgress = parseInt(message.payload)
                setProgress(currentProgress)
                if (currentProgress === 100) {
                    setStatus('File upload completed successfully!')
                }
            } else if (message.type === 'ready-next-chunk') {
                sendNextChunk()
            }
        }

        ws.current.onerror = (error) => {
            setStatus('WebSocket error')
            console.error('WebSocket error:', error)
        }

        ws.current.onclose = () => {
            setStatus('Disconnected from server')
        }

        return () => {
            ws.current?.close()
        }
    }, [])

    const sendNextChunk = () => {
        if (!fileRef.current || !ws.current) return

        const chunk = fileRef.current.slice(
            offsetRef.current,
            offsetRef.current + CHUNK_SIZE
        )

        const reader = new FileReader()
        reader.onload = () => {
            if (ws.current?.readyState === WebSocket.OPEN) {
                ws.current.send(JSON.stringify({
                    type: 'chunk',
                    payload: {
                        data: reader.result,
                        offset: offsetRef.current,
                        total: fileRef.current?.size || 0,
                        fileName: fileRef.current?.name
                    }
                }))
                offsetRef.current += chunk.size
            }
        }
        reader.readAsDataURL(chunk)
    }

    const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0]
        if (!file || !ws.current) return

        fileRef.current = file
        offsetRef.current = 0
        setProgress(0)

        // Start the upload by sending the first chunk
        sendNextChunk()
    }

    return (
        <div className="space-y-4 px-4 py-6 bg-white rounded-lg shadow-md">
            <div className="flex items-center justify-center">
                <label htmlFor="file-upload" className="flex items-center space-x-2">
                    <FaUpload className="text-blue-500" />
                    <span className="text-sm font-medium text-gray-700">Choose file</span>
                </label>
                <input
                    id="file-upload"
                    type="file"
                    onChange={handleFileChange}
                    className="hidden"
                />
            </div>

            {status && (
                <p className="text-sm text-gray-600">{status}</p>
            )}

            {progress > 0 && (
                <div className="w-full bg-gray-200 rounded-full h-3">
                    <div
                        className="bg-blue-600 h-3 rounded-full"
                        style={{ width: `${progress}%` }}
                    ></div>
                    <span className="text-sm text-gray-700 ml-2">{progress}%</span>
                </div>
            )}
        </div>
    )
}

export default FileUploader 