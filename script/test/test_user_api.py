#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
NovelAI 用户 API 测试脚本
功能：测试用户模块所有 HTTP 接口，包括注册、登录、获取信息、更新信息
"""

import requests
import json
import time
import sys
import socket
import os
from datetime import datetime
from colorama import Fore, Style, init

# 初始化 colorama
init(autoreset=True)

# 配置参数
API_URL = "http://localhost:8888"  # API服务地址
TIMEOUT = 5  # 请求超时时间(秒)
MAX_RETRY = 30  # 最大重试次数
RETRY_INTERVAL = 2  # 重试间隔(秒)

# 测试数据
timestamp = int(time.time())
test_data = {
    "username": f"testuser_{timestamp}",
    "password": "Test123456",
    "email": f"test_{timestamp}@example.com",
    "nickname": "测试用户"
}

# 用于存储测试过程中的数据
test_context = {
    "user_id": None,
    "token": None
}

def print_divider():
    """打印分隔线"""
    print(f"\n{Fore.YELLOW}{'━' * 80}{Style.RESET_ALL}\n")

def print_result(test_name, success, message=""):
    """
    打印测试结果
    
    Args:
        test_name: 测试名称
        success: 测试是否成功
        message: 附加信息
    """
    if success:
        print(f"{Fore.GREEN}✓ {test_name}: 成功 {message}{Style.RESET_ALL}")
    else:
        print(f"{Fore.RED}✗ {test_name}: 失败 {message}{Style.RESET_ALL}")
    return success

def wait_for_service():
    """等待 API 服务启动"""
    print("正在等待 API 服务启动...")
    
    for attempt in range(MAX_RETRY):
        try:
            response = requests.get(f"{API_URL}/ping", timeout=TIMEOUT)
            if response.status_code == 200:
                print(f"{Fore.GREEN}API 服务已启动并可访问{Style.RESET_ALL}")
                return True
        except (requests.RequestException, socket.error):
            pass
            
        print(f"等待 API 服务，尝试 {attempt+1}/{MAX_RETRY}...")
        time.sleep(RETRY_INTERVAL)
    
    print(f"{Fore.RED}错误: API 服务未能在预期时间内启动{Style.RESET_ALL}")
    return False

def pretty_print_response(response):
    """
    美化打印 HTTP 响应
    
    Args:
        response: requests 的响应对象
    """
    print("收到的原始响应:")
    try:
        content = response.json()
        print(json.dumps(content, ensure_ascii=False, indent=2))
    except ValueError:
        print(response.text)
    print("")
    
    return response

def test_register():
    """
    测试用户注册 API
    
    Returns:
        bool: 测试是否成功
    """
    print_divider()
    print(f"测试用户注册 API [POST /api/user/register]")
    print(f"请求数据: 用户名='{test_data['username']}', 密码='{test_data['password']}'")
    
    # 发送注册请求
    try:
        response = requests.post(
            f"{API_URL}/api/user/register",
            json={
                "username": test_data["username"],
                "password": test_data["password"],
                "nickname": test_data["nickname"],
                "email": test_data["email"]
            },
            timeout=TIMEOUT
        )
        
        pretty_print_response(response)
        
        # 确保响应是有效的 JSON
        try:
            result = response.json()
        except ValueError:
            return print_result("用户注册", False, "(响应不是有效的 JSON)")
        
        # 检查响应状态
        if response.status_code != 200:
            return print_result("用户注册", False, f"(HTTP状态码: {response.status_code})")
        
        # 检查业务状态码
        if "code" in result and result["code"] == 0:
            # 存储用户ID和令牌
            test_context["user_id"] = result.get("user_id")
            test_context["token"] = result.get("token")
            
            return print_result(
                "用户注册", 
                True, 
                f"(用户ID: {test_context['user_id']}, 令牌: {test_context['token'][:10]}...)"
            )
        elif "user_id" in result and "token" in result:
            # 有些接口可能没有code字段但返回了user_id和token
            test_context["user_id"] = result.get("user_id")
            test_context["token"] = result.get("token")
            
            return print_result(
                "用户注册", 
                True, 
                f"(用户ID: {test_context['user_id']}, 令牌: {test_context['token'][:10]}...)"
            )
        else:
            code = result.get("code", "未知")
            message = result.get("message", "无错误信息")
            return print_result("用户注册", False, f"(错误码: {code}, 消息: {message})")
            
    except requests.RequestException as e:
        return print_result("用户注册", False, f"(请求异常: {str(e)})")

def test_duplicate_register():
    """
    测试重复用户注册 API (应当失败)
    
    Returns:
        bool: 测试是否成功
    """
    print_divider()
    print(f"测试重复用户注册 API [POST /api/user/register] (预期失败)")
    
    # 发送重复的注册请求
    try:
        response = requests.post(
            f"{API_URL}/api/user/register",
            json={
                "username": test_data["username"],
                "password": test_data["password"],
                "nickname": test_data["nickname"],
                "email": test_data["email"]
            },
            timeout=TIMEOUT
        )
        
        pretty_print_response(response)
        
        # 确保响应是有效的 JSON
        try:
            result = response.json()
        except ValueError:
            return print_result("重复用户注册", False, "(响应不是有效的 JSON)")
        
        # 检查业务状态码 - 这里我们期望失败，错误码为 1001
        code = result.get("code", "未知")
        message = result.get("message", "无错误信息")
        
        if code == 1001 and "用户名已存在" in message:
            return print_result("重复用户注册", True, "(正确返回错误)")
        else:
            return print_result("重复用户注册", False, f"(未正确处理重复用户), 错误码: {code}, 消息: {message}")
            
    except requests.RequestException as e:
        return print_result("重复用户注册", False, f"(请求异常: {str(e)})")

def test_login():
    """
    测试用户登录 API
    
    Returns:
        bool: 测试是否成功
    """
    print_divider()
    print(f"测试用户登录 API [POST /api/user/login]")
    
    # 发送登录请求
    try:
        response = requests.post(
            f"{API_URL}/api/user/login",
            json={
                "username": test_data["username"],
                "password": test_data["password"]
            },
            timeout=TIMEOUT
        )
        
        pretty_print_response(response)
        
        # 确保响应是有效的 JSON
        try:
            result = response.json()
        except ValueError:
            return print_result("用户登录", False, "(响应不是有效的 JSON)")
        
        # 检查响应状态
        if response.status_code != 200:
            return print_result("用户登录", False, f"(HTTP状态码: {response.status_code})")
        
        # 检查业务状态码
        if "code" in result and result["code"] == 0:
            login_user_id = result.get("user_id")
            login_token = result.get("token")
            
            # 验证返回的信息与注册时相符
            if str(login_user_id) == str(test_context["user_id"]):
                test_context["token"] = login_token  # 更新令牌
                return print_result(
                    "用户登录", 
                    True, 
                    f"(用户ID: {login_user_id}, 令牌: {login_token[:10]}...)"
                )
            else:
                return print_result(
                    "用户登录", 
                    False, 
                    f"(用户ID不匹配, 期望: {test_context['user_id']}, 实际: {login_user_id})"
                )
        elif "user_id" in result and "token" in result:
            login_user_id = result.get("user_id")
            login_token = result.get("token")
            
            # 验证返回的信息与注册时相符
            if str(login_user_id) == str(test_context["user_id"]):
                test_context["token"] = login_token  # 更新令牌
                return print_result(
                    "用户登录", 
                    True, 
                    f"(用户ID: {login_user_id}, 令牌: {login_token[:10]}...)"
                )
            else:
                return print_result(
                    "用户登录", 
                    False, 
                    f"(用户ID不匹配, 期望: {test_context['user_id']}, 实际: {login_user_id})"
                )
        else:
            code = result.get("code", "未知")
            message = result.get("message", "无错误信息")
            return print_result("用户登录", False, f"(错误码: {code}, 消息: {message})")
            
    except requests.RequestException as e:
        return print_result("用户登录", False, f"(请求异常: {str(e)})")

def test_wrong_password_login():
    """
    测试错误密码登录 API (应当失败)
    
    Returns:
        bool: 测试是否成功
    """
    print_divider()
    print(f"测试错误密码登录 API [POST /api/user/login] (预期失败)")
    
    # 发送错误密码登录请求
    try:
        response = requests.post(
            f"{API_URL}/api/user/login",
            json={
                "username": test_data["username"],
                "password": "WrongPassword123"
            },
            timeout=TIMEOUT
        )
        
        pretty_print_response(response)
        
        # 确保响应是有效的 JSON
        try:
            result = response.json()
        except ValueError:
            return print_result("错误密码登录", False, "(响应不是有效的 JSON)")
        
        # 检查业务状态码 - 这里我们期望失败，错误码为 1002
        code = result.get("code", "未知")
        message = result.get("message", "无错误信息")
        
        if code == 1002 and "用户名或密码错误" in message:
            return print_result("错误密码登录", True, "(正确返回错误)")
        else:
            return print_result("错误密码登录", False, f"(未正确处理错误密码), 错误码: {code}, 消息: {message}")
            
    except requests.RequestException as e:
        return print_result("错误密码登录", False, f"(请求异常: {str(e)})")

def test_get_user_info():
    """
    测试获取用户信息 API
    
    Returns:
        bool: 测试是否成功
    """
    print_divider()
    print(f"测试获取用户信息 API [GET /api/user/info]")
    
    if not test_context["user_id"] or not test_context["token"]:
        return print_result("获取用户信息", False, "(缺少用户ID或令牌)")
    
    # 发送获取用户信息请求
    try:
        response = requests.get(
            f"{API_URL}/api/user/info?user_id={test_context['user_id']}",
            headers={"Authorization": f"Bearer {test_context['token']}"},
            timeout=TIMEOUT
        )
        
        pretty_print_response(response)
        
        # 确保响应是有效的 JSON
        try:
            result = response.json()
        except ValueError:
            return print_result("获取用户信息", False, "(响应不是有效的 JSON)")
        
        # 检查响应状态
        if response.status_code != 200:
            return print_result("获取用户信息", False, f"(HTTP状态码: {response.status_code})")
        
        # 适配实际API响应格式，检查成功消息或代码
        success = False
        if ("code" in result and result["code"] == 0) or \
           ("message" in result and "成功" in result.get("message", "")):
            success = True
            
        if success:
            # 优先检查实际响应中的user字段(实际API格式)，然后再尝试data字段(测试预期格式)
            user_data = None
            if "user" in result:
                user_data = result["user"]
            elif "data" in result:
                user_data = result["data"]
                
            if user_data:
                returned_username = user_data.get("username")
                
                if returned_username == test_data["username"]:
                    return print_result("获取用户信息", True, f"(用户名: {returned_username})")
                else:
                    return print_result(
                        "获取用户信息", 
                        False, 
                        f"(用户名不匹配: {returned_username} 应为 {test_data['username']})"
                    )
            else:
                return print_result("获取用户信息", False, "(响应中缺少user或data字段)")
        else:
            code = result.get("code", "未知")
            message = result.get("message", "无错误信息")
            return print_result("获取用户信息", False, f"(错误码: {code}, 消息: {message})")
            
    except requests.RequestException as e:
        return print_result("获取用户信息", False, f"(请求异常: {str(e)})")

def test_update_user():
    """
    测试更新用户信息 API
    
    Returns:
        bool: 测试是否成功
    """
    print_divider()
    print(f"测试更新用户信息 API [PUT /api/user/update]")
    
    if not test_context["user_id"] or not test_context["token"]:
        return print_result("更新用户信息", False, "(缺少用户ID或令牌)")
    
    # 新的昵称和头像
    new_nickname = f"更新后的测试用户_{timestamp}"
    new_avatar = f"http://example.com/avatar_{timestamp}.jpg"
    
    # 发送更新用户信息请求
    try:
        response = requests.put(
            f"{API_URL}/api/user/update",
            headers={
                "Content-Type": "application/json",
                "Authorization": f"Bearer {test_context['token']}"
            },
            json={
                "user_id": test_context["user_id"],
                "nickname": new_nickname,
                "avatar": new_avatar
            },
            timeout=TIMEOUT
        )
        
        pretty_print_response(response)
        
        # 确保响应是有效的 JSON
        try:
            result = response.json()
        except ValueError:
            return print_result("更新用户信息", False, "(响应不是有效的 JSON)")
        
        # 检查响应状态
        if response.status_code != 200:
            return print_result("更新用户信息", False, f"(HTTP状态码: {response.status_code})")
        
        # 检查业务状态码
        update_success = False
        if "code" in result and result["code"] == 0:
            update_success = True
        elif "message" in result and "成功" in result.get("message", ""):
            # 接口可能没有返回code但有成功的message
            update_success = True
            
        if update_success:
            print_result("更新用户信息", True)
            
            # 验证更新后的信息
            try:
                verify_response = requests.get(
                    f"{API_URL}/api/user/info?user_id={test_context['user_id']}",
                    headers={"Authorization": f"Bearer {test_context['token']}"},
                    timeout=TIMEOUT
                )
                
                print("验证更新时收到的响应:")
                try:
                    verify_result = verify_response.json()
                    print(json.dumps(verify_result, ensure_ascii=False, indent=2))
                except ValueError:
                    print(verify_response.text)
                print("")
                
                # 检查是否为有效JSON
                if verify_response.status_code != 200:
                    return print_result("验证更新结果", False, f"(HTTP状态码: {verify_response.status_code})")
                
                # 获取更新后的昵称，适配实际API响应格式
                updated_nickname = None
                # 优先检查user字段（实际API格式）
                if "user" in verify_result:
                    updated_nickname = verify_result["user"].get("nickname")
                # 如果没有找到，再检查data字段（测试预期格式）
                elif "data" in verify_result:
                    updated_nickname = verify_result["data"].get("nickname")
                
                if updated_nickname == new_nickname:
                    return print_result("验证更新结果", True, f"(昵称已更新为: {updated_nickname})")
                else:
                    return print_result(
                        "验证更新结果", 
                        False, 
                        f"(期望昵称: {new_nickname}, 实际昵称: {updated_nickname or '未找到'})"
                    )
                
            except requests.RequestException as e:
                return print_result("验证更新结果", False, f"(请求异常: {str(e)})")
        else:
            code = result.get("code", "未知")
            message = result.get("message", "无错误信息")
            return print_result("更新用户信息", False, f"(错误码: {code}, 消息: {message})")
            
    except requests.RequestException as e:
        return print_result("更新用户信息", False, f"(请求异常: {str(e)})")

def run_tests():
    """运行所有测试"""
    test_results = []
    
    # 等待API服务启动
    if not wait_for_service():
        return False
    
    # 运行测试用例
    test_results.append(test_register())
    test_results.append(test_duplicate_register())
    test_results.append(test_login())
    test_results.append(test_wrong_password_login())
    test_results.append(test_get_user_info())
    test_results.append(test_update_user())
    
    # 统计结果
    success_count = sum(1 for result in test_results if result)
    total_count = len(test_results)
    
    print_divider()
    print(f"测试完成: {success_count}/{total_count} 通过")
    
    # 如果有失败的测试，返回非零状态码
    return all(test_results)

def main():
    """主函数"""
    try:
        print(f"{Fore.CYAN}开始 NovelAI 用户 API 测试...{Style.RESET_ALL}")
        result = run_tests()
        
        if result:
            print(f"{Fore.GREEN}所有测试通过!{Style.RESET_ALL}")
            return 0
        else:
            print(f"{Fore.YELLOW}测试中有失败项，请检查上面的输出。{Style.RESET_ALL}")
            return 1
    except KeyboardInterrupt:
        print(f"{Fore.RED}测试被用户中断{Style.RESET_ALL}")
        return 130
    except Exception as e:
        print(f"{Fore.RED}测试过程中发生未处理的异常: {str(e)}{Style.RESET_ALL}")
        import traceback
        traceback.print_exc()
        return 1

if __name__ == "__main__":
    sys.exit(main())
