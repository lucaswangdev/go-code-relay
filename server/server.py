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
import re
from anthropic import Anthropic

# ============== 配置 ==============
API_KEY = os.environ.get("ANTHROPIC_API_KEY", "")
BASE_URL = os.environ.get("ANTHROPIC_BASE_URL", "https://api.minimaxi.com/anthropic")
HOST = os.environ.get("RELAY_HOST", "0.0.0.0")
PORT = int(os.environ.get("RELAY_PORT", "8765"))
MODEL = os.environ.get("ANTHROPIC_MODEL", "MiniMax-M2.7")
MAX_CONTEXT_TOKENS = int(os.environ.get("MAX_CONTEXT_TOKENS", "150000"))
# =================================

if not API_KEY:
    print("错误: 请设置 ANTHROPIC_API_KEY 环境变量")
    sys.exit(1)

client = Anthropic(api_key=API_KEY, base_url=BASE_URL)

# 存储每个客户端的数据
class ClientSession:
    def __init__(self):
        self.messages = []
        self.modified_files = set()  # 追踪修改的文件
        self.deleted_files = set()   # 追踪删除的文件

client_sessions = {}

def approx_tokens(text: str) -> int:
    """估算 token 数量（中文约 1 token = 1-2 字符）"""
    chinese_chars = len(re.findall(r'[\u4e00-\u9fff]', text))
    other_chars = len(text) - chinese_chars
    return chinese_chars + other_chars // 4

def estimate_tokens(messages: list) -> int:
    """估算消息列表的总 token 数"""
    total = 0
    for m in messages:
        if content := m.get("content"):
            if isinstance(content, str):
                total += approx_tokens(content)
    return total

def snip_tool_outputs(messages: list) -> bool:
    """压缩过长的 tool 输出"""
    changed = False
    for m in messages:
        if m.get("role") != "tool":
            continue
        content = m.get("content", "")
        if not isinstance(content, str) or len(content) <= 1500:
            continue
        lines = content.split("\n")
        if len(lines) <= 6:
            continue
        snipped = "\n".join(lines[:3])
        snipped += f"\n... ({len(lines)} lines, snipped) ...\n"
        snipped += "\n".join(lines[-3:])
        m["content"] = snipped
        changed = True
    return changed

async def send_json(ws, data):
    """发送 JSON 消息"""
    try:
        await ws.send(json.dumps(data))
    except Exception as e:
        print(f"发送失败: {e}")

async def relay_handler(websocket):
    """处理客户端连接"""
    client_id = id(websocket)
    client_sessions[client_id] = ClientSession()
    session = client_sessions[client_id]
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
                session.messages = []
                session.modified_files = set()
                session.deleted_files = set()
                await send_json(websocket, {"type": "cleared"})
                print(f"[客户端 {client_id}] 历史已清空")

            elif msg_type == "message":
                user_input = message.get("content", "")
                print(f"[客户端 {client_id}] 收到消息: {user_input[:50]}...")

                session.messages.append({"role": "user", "content": user_input})

                # Context 压缩逻辑
                current_tokens = estimate_tokens(session.messages)
                if current_tokens > MAX_CONTEXT_TOKENS * 0.7:
                    print(f"[客户端 {client_id}] Context 压缩触发 ({current_tokens} tokens)")
                    if snip_tool_outputs(session.messages):
                        print(f"[客户端 {client_id}] Tool 输出已压缩")

                try:
                    with client.messages.stream(
                        model=MODEL,
                        max_tokens=2048,
                        messages=session.messages
                    ) as stream:
                        full_response = ""
                        for text in stream.text_stream:
                            await send_json(websocket, {"type": "chunk", "content": text})
                            full_response += text

                        session.messages.append({"role": "assistant", "content": full_response})

                    await send_json(websocket, {"type": "done"})
                    print(f"[客户端 {client_id}] 响应完成 ({len(full_response)} 字符)")

                except Exception as e:
                    error_msg = f"API 调用失败: {e}"
                    print(f"[客户端 {client_id}] {error_msg}")
                    await send_json(websocket, {"type": "error", "content": error_msg})

            elif msg_type == "diff":
                # 返回修改的文件列表
                diff_info = {
                    "type": "diff",
                    "modified": list(session.modified_files),
                    "deleted": list(session.deleted_files)
                }
                await send_json(websocket, diff_info)

            elif msg_type == "file_changed":
                # 追踪文件变化
                change_type = message.get("change_type", "")  # "modified" or "deleted"
                file_path = message.get("file_path", "")
                if change_type == "modified":
                    session.modified_files.add(file_path)
                elif change_type == "deleted":
                    session.deleted_files.add(file_path)
                    session.modified_files.discard(file_path)

            elif msg_type == "ping":
                await send_json(websocket, {"type": "pong"})

            else:
                await send_json(websocket, {"type": "error", "content": f"未知消息类型: {msg_type}"})

    except websockets.exceptions.ConnectionClosed:
        print(f"[客户端 {client_id}] 连接关闭")
    except Exception as e:
        print(f"[客户端 {client_id}] 错误: {e}")
    finally:
        if client_id in client_sessions:
            del client_sessions[client_id]
        print(f"[客户端 {client_id}] 断开连接")

async def main():
    print("=" * 50)
    print("Claude Relay Server")
    print("=" * 50)
    print(f"API Key: {API_KEY[:20]}..." if len(API_KEY) > 20 else "未设置")
    print(f"Base URL: {BASE_URL}")
    print(f"Model: {MODEL}")
    print(f"Max Context: {MAX_CONTEXT_TOKENS} tokens")
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
