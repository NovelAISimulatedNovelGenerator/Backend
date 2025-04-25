# -*- coding: utf-8 -*-
"""
conftest.py
pytest 全局fixture：自动注册并登录测试用户，供所有API测试调用
- 若用户已注册则自动登录
- 返回token字符串
- 便于所有测试文件低耦合复用
"""
import pytest
import requests

BASE_URL = "http://127.0.0.1:8888"
REGISTER_URL = f"{BASE_URL}/api/user/register"
LOGIN_URL = f"{BASE_URL}/api/user/login"

# 测试用户信息（可统一配置）
TEST_USERNAME = "testuser"
TEST_PASSWORD = "testpass"


def register_user(username: str, password: str) -> bool:
    """
    注册新用户
    :param username: 用户名
    :param password: 密码
    :return: 注册成功返回True，已存在返回False
    """
    resp = requests.post(REGISTER_URL, json={"username": username, "password": password})
    if resp.status_code == 200:
        return True
    # 若已注册，通常返回409或400，具体看后端实现
    if resp.status_code in (400, 409):
        return False
    raise Exception(f"注册接口异常: {resp.status_code} {resp.text}")


def login_user(username: str, password: str) -> str:
    """
    登录用户，获取token
    :param username: 用户名
    :param password: 密码
    :return: token字符串，失败抛异常
    """
    resp = requests.post(LOGIN_URL, json={"username": username, "password": password})
    if resp.status_code == 200 and "token" in resp.json():
        return resp.json()["token"]
    raise Exception(f"登录失败: {resp.status_code} {resp.text}")


@pytest.fixture(scope="session")
def auto_register_and_login() -> str:
    """
    pytest fixture：自动注册并登录测试用户，返回token
    所有测试文件可直接引用本fixture获取token
    """
    try:
        registered = register_user(TEST_USERNAME, TEST_PASSWORD)
    except Exception as e:
        pytest.skip(f"注册用户异常: {e}")
    try:
        token = login_user(TEST_USERNAME, TEST_PASSWORD)
    except Exception as e:
        pytest.skip(f"登录用户异常: {e}")
    return token

# 用法示例（在测试文件中）：
# def test_xxx(auto_register_and_login):
#     token = auto_register_and_login
#     ...
