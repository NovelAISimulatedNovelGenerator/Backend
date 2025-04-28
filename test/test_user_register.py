# -*- coding: utf-8 -*-
"""
test_user_register.py
用户注册接口（/api/user/register）API单元测试

- 覆盖正常注册、用户名已存在、参数缺失等情况
- 需先启动后端服务
- 依赖 pytest fixture auto_register_and_login，自动注册并登录测试用户
"""
import requests
import pytest
import random
import string

BASE_URL = "http://127.0.0.1:8888"
REGISTER_URL = f"{BASE_URL}/api/user/register"

def random_username(length=8):
    """生成随机用户名，避免冲突"""
    return "testuser_" + ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

class TestUserRegister:
    """
    /api/user/register 用户注册接口测试
    """
    def test_register_success(self):
        """
        正常注册新用户
        """
        username = random_username()
        password = "testpass"
        resp = requests.post(REGISTER_URL, json={"username": username, "password": password})
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200
        assert data["message"] == "注册成功"
        assert "user_id" in data
        assert "token" in data

    def test_register_user_exists(self, auto_register_and_login):
        """
        已存在用户名注册，应返回用户名已存在
        """
        # 使用 fixture 注册的用户名
        from conftest import TEST_USERNAME, TEST_PASSWORD
        resp = requests.post(REGISTER_URL, json={"username": TEST_USERNAME, "password": TEST_PASSWORD})
        assert resp.status_code == 200
        data = resp.json()
        # 约定 code=1001 表示用户名已存在
        assert data["code"] == 1001
        assert "已存在" in data["message"]

    def test_register_missing_params(self):
        """
        缺少必要参数，接口应返回错误
        """
        resp = requests.post(REGISTER_URL, json={"username": "abc"})  # 缺 password
        assert resp.status_code in (400, 200)
        data = resp.json()
        # 约定 code=400 表示参数错误
        assert data["code"] in (400, 500)

# 测试流程说明：
# 1. test_register_success：注册新用户，断言返回内容正确
# 2. test_register_user_exists：注册已存在用户名，应返回用户名已存在
# 3. test_register_missing_params：缺失参数，接口应返回参数错误
# 运行方法：pytest test_user_register.py
