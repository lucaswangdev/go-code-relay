#!/usr/bin/env python3
"""
Claude Relay Server - 运行在 B 电脑
接收 A 电脑的请求，转发给 Claude API，返回流式响应
"""

import asyncio
import websockets
import json
import os
import sys
from anthropic import Anthropic

# ============== 配置 ==============
# 从环境变量读取 API Key（不要硬编码！）
API_KEY = os.environ.get("ANTHROPIC_API_KEY", "")
BASE_URL = os.environ.get("ANTHROPIC_BASE_URL", "https://api.minimaxi.com/anthropic")
HOST = os.environ.get("RELAY_HOST", "0.0.0.0")
PORT = int(os.environ.get("RELAY_PORT", "8765"))
MODEL = os.environ.get("ANTHROPIC_MODEL", "MiniMax-M2.7")
# =================================

if not API_KEY:
    print("错误: 请设置 ANTHROPIC_API_KEY 环境变量")
    print("示例: export ANTHROPIC_API_KEY=sk-ant-...")
    sys.exit(1)

client = Anthropic(api_key=API_KEY, base_url=BASE_URL)

# 存储每个客户端的对话历史
client_histories = {}

async def send_json(ws, data):
    """发送 JSON 消息"""
    try:
        await ws.send(json.dumps(data))
    except Exception as e:
        print(f"发送失败: {e}")

async def relay_handler(websocket):
    """处理客户端连接"""
    client_id = id(websocket)
    client_histories[client_id] = []
    print(f"[客户端 {client_id}] 连接")

    try:
        async for raw_message in websocket:
            try:
                message = json.loads(raw_message)
            except json.JSONDecodeError:
                await send_json(websocket, {"type": "error", "content": "无效的 JSON 格式"})
                continue

            msg_type = message.get("type", "")

            if msg_type == "clear":
                client_histories[client_id] = []
                await send_json(websocket, {"type": "cleared"})
                print(f"[客户端 {client_id}] 历史已清空")

            elif msg_type == "message":
                user_input = message.get("content", "")
                print(f"[客户端 {client_id}] 收到消息: {user_input[:50]}...")

                messages = client_histories[client_id] + [{"role": "user", "content": user_input}]

                try:
                    with client.messages.stream(
                        model=MODEL,
                        max_tokens=2048,
                        messages=messages
                    ) as stream:
                        full_response = ""
                        for text in stream.text_stream:
                            await send_json(websocket, {"type": "chunk", "content": text})
                            full_response += text

                        client_histories[client_id].append({"role": "user", "content": user_input})
                        client_histories[client_id].append({"role": "assistant", "content": full_response})

                    await send_json(websocket, {"type": "done"})
                    print(f"[客户端 {client_id}] 响应完成 ({len(full_response)} 字符)")

                except Exception as e:
                    error_msg = f"API 调用失败: {e}"
                    print(f"[客户端 {client_id}] {error_msg}")
                    await send_json(websocket, {"type": "error", "content": error_msg})

            elif msg_type == "ping":
                await send_json(websocket, {"type": "pong"})
            else:
                await send_json(websocket, {"type": "error", "content": f"未知消息类型: {msg_type}"})

    except websockets.exceptions.ConnectionClosed:
        print(f"[客户端 {client_id}] 连接关闭")
    except Exception as e:
        print(f"[客户端 {client_id}] 错误: {e}")
    finally:
        if client_id in client_histories:
            del client_histories[client_id]
        print(f"[客户端 {client_id}] 断开连接")

async def main():
    print("=" * 50)
    print("Claude Relay Server")
    print("=" * 50)
    print(f"API Key: {API_KEY[:20]}..." if len(API_KEY) > 20 else "未设置")
    print(f"Base URL: {BASE_URL}")
    print(f"Model: {MODEL}")
    print(f"监听地址: ws://{HOST}:{PORT}")
    print("=" * 50)

    async with websockets.serve(relay_handler, HOST, PORT):
        print("服务器已启动，按 Ctrl+C 停止")
        await asyncio.Future()

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n服务器已停止")
