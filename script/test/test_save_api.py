#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
NovelAI 保存模块 API 测试脚本
功能：测试保存(save)模块所有 HTTP 接口，包括创建、获取、更新、删除、列表
"""

import requests
import json
import time
import sys
import socket
from colorama import Fore, Style, init

# 初始化 colorama
init(autoreset=True)

API_URL = "http://localhost:8888"
TIMEOUT = 5
MAX_RETRY = 30
RETRY_INTERVAL = 2

def print_divider():
    print(f"\n{Fore.YELLOW}{'━' * 80}{Style.RESET_ALL}\n")

def print_result(test_name, success, message=""):
    if success:
        print(f"{Fore.GREEN}✓ {test_name}: 成功 {message}{Style.RESET_ALL}")
    else:
        print(f"{Fore.RED}✗ {test_name}: 失败 {message}{Style.RESET_ALL}")
    return success

def wait_for_service():
    print("正在等待 API 服务启动...")
    for attempt in range(MAX_RETRY):
        try:
            resp = requests.get(f"{API_URL}/ping", timeout=TIMEOUT)
            if resp.status_code == 200:
                print(f"{Fore.GREEN}API 服务已启动并可访问{Style.RESET_ALL}")
                return True
        except (requests.RequestException, socket.error):
            pass
        print(f"等待 API 服务，尝试 {attempt+1}/{MAX_RETRY}...")
        time.sleep(RETRY_INTERVAL)
    print(f"{Fore.RED}错误: API 服务未能在预期时间内启动{Style.RESET_ALL}")
    return False

def pretty_print_response(response):
    print("收到的原始响应:")
    try:
        content = response.json()
        print(json.dumps(content, ensure_ascii=False, indent=2))
    except ValueError:
        print(response.text)
    print("")
    return response

# 测试数据
# 用户信息配置（如需更换测试用户请修改此处）
USERNAME = "test_save_user"
PASSWORD = "test_save_pwd"
NICKNAME = "保存测试用户"
EMAIL = f"save_test_{int(time.time())}@test.com"
user_id = None  # 动态获取
save_id = None
TOKEN = None

import json
test_save_data = {
    "save_name": f"测试存档_{int(time.time())}",
    "save_description": "单元测试用存档描述",
    "save_data": json.dumps({"chapter": 1, "text": "测试内容"}),
    "save_type": "1"  # 必须为字符串
}

def get_token_and_userid():
    """
    注册新用户，若已存在则自动登录，返回token和user_id
    增加详细日志，便于排查注册与登录流程问题。
    """
    global TOKEN, user_id
    # 注册
    reg_url = f"{API_URL}/api/user/register"
    reg_payload = {"username": USERNAME, "password": PASSWORD, "nickname": NICKNAME, "email": EMAIL}
    print("[LOG] 注册请求 payload:", reg_payload)
    resp = requests.post(reg_url, json=reg_payload, timeout=TIMEOUT)
    print("[LOG] 注册响应 status:", resp.status_code)
    try:
        data = resp.json()
        print("[LOG] 注册响应内容:", data)
        if data.get("code") == 200:
            user_id = data.get("user_id")
            print(f"[LOG] 注册成功，user_id={user_id}")
        elif data.get("code") == 1001:  # 用户已存在
            print("[LOG] 用户已存在，准备直接登录")
        else:
            print_result("注册用户", False, data.get("message", "注册失败"))
            sys.exit(1)
    except Exception as e:
        print_result("注册用户", False, str(e))
        sys.exit(1)
    # 登录
    login_url = f"{API_URL}/api/user/login"
    login_payload = {"username": USERNAME, "password": PASSWORD}
    print("[LOG] 登录请求 payload:", login_payload)
    resp = requests.post(login_url, json=login_payload, timeout=TIMEOUT)
    print("[LOG] 登录响应 status:", resp.status_code)
    try:
        data = resp.json()
        print("[LOG] 登录响应内容:", data)
        if data.get("code") == 200 and data.get("token"):
            TOKEN = data["token"]
            user_id = data.get("user_id")
            print(f"[LOG] 登录成功，user_id={user_id}，token={TOKEN}")
            print_result("登录用户", True)
        else:
            print_result("登录用户", False, data.get("message", "登录失败"))
            sys.exit(1)
    except Exception as e:
        print_result("登录用户", False, str(e))
        sys.exit(1)

def get_auth_headers():
    """
    获取带 JWT 的请求头
    """
    return {"Authorization": f"Bearer {TOKEN}"}

def test_create_save():
    print_divider()
    print("测试: 创建保存接口 (/api/save/create)")
    url = f"{API_URL}/api/save/create"
    payload = test_save_data.copy()
    payload["user_id"] = user_id
    print(f"[LOG] 创建保存请求 URL: {url}")
    print(f"[LOG] 创建保存请求 payload: {payload}")
    try:
        resp = requests.post(url, json=payload, timeout=TIMEOUT, headers=get_auth_headers())
        print(f"[LOG] 创建保存响应 status: {resp.status_code}")
        pretty_print_response(resp)
        data = resp.json()
        print(f"[LOG] 创建保存响应内容: {data}")
        if data.get("code") == 200 and data.get("save_id"):
            global save_id
            save_id = data["save_id"]
            return print_result("创建保存", True)
        else:
            return print_result("创建保存", False, data.get("message", "无返回信息"))
    except Exception as e:
        print(f"[LOG] 创建保存异常: {str(e)}")
        return print_result("创建保存", False, str(e))

def test_get_save():
    print_divider()
    print("测试: 获取保存接口 (/api/save/get)")
    url = f"{API_URL}/api/save/get"
    params = {"user_id": user_id, "save_id": save_id}
    print(f"[LOG] 获取保存请求 URL: {url}")
    print(f"[LOG] 获取保存请求 params: {params}")
    try:
        resp = requests.get(url, params=params, timeout=TIMEOUT, headers=get_auth_headers())
        print(f"[LOG] 获取保存响应 status: {resp.status_code}")
        pretty_print_response(resp)
        data = resp.json()
        print(f"[LOG] 获取保存响应内容: {data}")
        if data.get("code") == 200 and data.get("save"):
            return print_result("获取保存", True)
        else:
            return print_result("获取保存", False, data.get("message", "无返回信息"))
    except Exception as e:
        print(f"[LOG] 获取保存异常: {str(e)}")
        return print_result("获取保存", False, str(e))

def test_update_save():
    print_divider()
    print("测试: 更新保存接口 (/api/save/update)")
    url = f"{API_URL}/api/save/update"
    payload = test_save_data.copy()
    payload["user_id"] = user_id
    payload["save_id"] = save_id
    payload["save_description"] = "已更新描述"
    payload["save_data"] = json.dumps({"chapter": 1, "text": "测试内容"})
    print(f"[LOG] 更新保存请求 URL: {url}")
    print(f"[LOG] 更新保存请求 payload: {payload}")
    try:
        resp = requests.put(url, json=payload, timeout=TIMEOUT, headers=get_auth_headers())
        print(f"[LOG] 更新保存响应 status: {resp.status_code}")
        pretty_print_response(resp)
        data = resp.json()
        print(f"[LOG] 更新保存响应内容: {data}")
        if data.get("code") == 200:
            return print_result("更新保存", True)
        else:
            return print_result("更新保存", False, data.get("Message", "无返回信息"))
    except Exception as e:
        print(f"[LOG] 更新保存异常: {str(e)}")
        return print_result("更新保存", False, str(e))

def test_delete_save():
    print_divider()
    print("测试: 删除保存接口 (/api/save/delete)")
    url = f"{API_URL}/api/save/delete"
    params = {"user_id": user_id, "save_id": save_id}
    print(f"[LOG] 删除保存请求 URL: {url}")
    print(f"[LOG] 删除保存请求 params: {params}")
    try:
        resp = requests.delete(url, params=params, timeout=TIMEOUT, headers=get_auth_headers())
        print(f"[LOG] 删除保存响应 status: {resp.status_code}")
        pretty_print_response(resp)
        data = resp.json()
        print(f"[LOG] 删除保存响应内容: {data}")
        if data.get("code") == 200:
            return print_result("删除保存", True)
        else:
            return print_result("删除保存", False, data.get("Message", "无返回信息"))
    except Exception as e:
        print(f"[LOG] 删除保存异常: {str(e)}")
        return print_result("删除保存", False, str(e))

def test_list_saves():
    print_divider()
    print("测试: 保存列表接口 (/api/save/list)")
    url = f"{API_URL}/api/save/list"
    params = {"user_id": user_id, "page": 1, "page_size": 10}
    print(f"[LOG] 保存列表请求 URL: {url}")
    print(f"[LOG] 保存列表请求 params: {params}")
    try:
        resp = requests.get(url, params=params, timeout=TIMEOUT, headers=get_auth_headers())
        print(f"[LOG] 保存列表响应 status: {resp.status_code}")
        pretty_print_response(resp)
        data = resp.json()
        print(f"[LOG] 保存列表响应内容: {data}")
        if data.get("code") == 200 and isinstance(data.get("saves"), list):
            return print_result("保存列表", True)
        else:
            return print_result("保存列表", False, data.get("message", "无返回信息"))
    except Exception as e:
        print(f"[LOG] 保存列表异常: {str(e)}")
        return print_result("保存列表", False, str(e))

def run_tests():
    if not wait_for_service():
        sys.exit(1)
    get_token_and_userid()  # 自动注册/登录并获取token和user_id
    all_passed = True
    if not test_create_save():
        all_passed = False
    if not test_get_save():
        all_passed = False
    if not test_update_save():
        all_passed = False
    if not test_list_saves():
        all_passed = False
    if not test_delete_save():
        all_passed = False
    print_divider()
    print(f"所有保存模块接口测试 {'通过' if all_passed else '未通过'}")
    return all_passed

if __name__ == "__main__":
    run_tests()
