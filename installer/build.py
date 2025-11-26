import subprocess
import time
import os
import logging


def build_installer():
    """
    运行 makensis 编译 installer.nsi 并输出执行时间。
    """
    nsis_script_path = "installer.nsi"
    command = ["makensis", nsis_script_path]

    # 检查 NSIS 脚本文件是否存在
    if not os.path.exists(nsis_script_path):
        logging.info(f"错误: 找不到 NSIS 脚本 '{nsis_script_path}'")
        return

    logging.info(f"正在运行: {' '.join(command)}")
    logging.info(f"{...}")

    start_time = time.monotonic()

    try:
        # 在 Windows 中，控制台输出通常使用 'cp936' (GBK) 编码
        result = subprocess.run(
            command, capture_output=True, text=True, encoding="cp936", errors="replace"
        )

        # 打印 makensis 的输出
        if result.stdout:
            logging.info("--- makensis 输出 ---")
            logging.info(f"{...}")
        if result.stderr:
            logging.info("--- makensis 错误 ---")
            logging.info(f"{...}")

        logging.info(f"{...}")

        if result.returncode == 0:
            logging.info("✅ 编译成功。")
        else:
            logging.info(f"❌ 编译失败，返回代码: {result.returncode}")

    except FileNotFoundError:
        logging.info("错误: 'makensis' 命令未找到。")
        logging.info(
            "请确保 NSIS 已安装，并且 'makensis.exe' 位于系统的 PATH 环境变量中。"
        )
        return
    except Exception as e:
        logging.exception(f"发生未知错误: {e}")
        return
    finally:
        end_time = time.monotonic()
        duration = end_time - start_time
        logging.info(f"⏱️  执行时间: {duration:.3f} 秒")


if __name__ == "__main__":
    build_installer()
