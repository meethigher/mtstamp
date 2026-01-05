## mtstamp

**mtstamp** 是一个使用 Go 1.20.14 开发的跨平台命令行工具，用于记录和恢复文件的修改时间。

### 功能

- **log**: 遍历指定目录下所有文件，记录每个文件的相对路径和修改时间到 `config.mtime`
- **back**: 根据 `config.mtime` 中记录的信息，恢复目录下所有文件的修改时间

### 支持的平台

- Windows
- Linux

### 编译

在项目根目录执行（需要安装 Go 1.20.14 或兼容版本）：

**Windows:**
```bash
go build -o mtstamp.exe
```

**Linux:**
```bash
go build -o mtstamp
```

编译成功后，将生成对应的可执行文件。

### 使用说明

#### log 命令 - 记录文件修改时间

**语法：**
```bash
mtstamp log <目录路径> [保存目录]
```

**参数说明：**
- `<目录路径>`: 要遍历的目录，**必须使用绝对路径**
- `[保存目录]`: config.mtime 的保存目录（可选），**必须使用绝对路径**
  - 如果不指定此参数，config.mtime 将保存在**运行命令时所在的目录**（当前工作目录）

**示例：**

```bash
# Windows: 不指定保存目录，config.mtime 保存在当前目录
mtstamp.exe log D:\work\project

# Windows: 指定保存目录
mtstamp.exe log D:\work\project D:\backup

# Linux: 不指定保存目录，config.mtime 保存在当前目录
./mtstamp log /home/user/project

# Linux: 指定保存目录
./mtstamp log /home/user/project /home/user/backup
```

#### back 命令 - 恢复文件修改时间

**语法：**
```bash
mtstamp back <目录路径> [config.mtime所在目录]
```

**参数说明：**
- `<目录路径>`: 要恢复文件时间的目录，**必须使用绝对路径**
- `[config.mtime所在目录]`: config.mtime 文件所在的目录（可选），**必须使用绝对路径**
  - 如果不指定此参数，则使用**运行命令时所在的目录**（当前工作目录）下的 `config.mtime`

**示例：**

```bash
# Windows: 不指定config.mtime所在目录，使用当前目录下的config.mtime
mtstamp.exe back D:\work\project

# Windows: 指定config.mtime所在目录
mtstamp.exe back D:\work\project D:\backup

# Linux: 不指定config.mtime所在目录，使用当前目录下的config.mtime
./mtstamp back /home/user/project

# Linux: 指定config.mtime所在目录
./mtstamp back /home/user/project /home/user/backup
```

### 详细说明

#### log 命令的工作流程

1. **参数验证**
   - 检查 `<目录路径>` 是否存在且为目录
   - 验证 `<目录路径>` 必须是绝对路径
   - 如果指定了 `[保存目录]`，验证其必须是绝对路径，并确保目录存在（不存在则自动创建）

2. **文件遍历**
   - 递归遍历 `<目录路径>` 下的所有文件（不包括子目录本身）
   - 跳过 `config.mtime` 文件本身（无论它保存在哪里）

3. **记录格式**
   - 每个文件记录一行，格式为：`<相对路径>\t<yyyyMMddHHmmss格式的时间>`
   - 相对路径是相对于 `<目录路径>` 的路径，使用正斜杠 `/` 分隔（跨平台兼容）
   - 时间格式为 yyyyMMddHHmmss（14位数字），例如：`20231225143045` 表示 2023年12月25日14点30分45秒

4. **保存位置**
   - 如果指定了 `[保存目录]`，`config.mtime` 保存在该目录下
   - 如果未指定 `[保存目录]`，`config.mtime` 保存在**运行命令时所在的目录**（当前工作目录）

#### back 命令的工作流程

1. **参数验证**
   - 检查 `<目录路径>` 是否存在且为目录
   - 验证 `<目录路径>` 必须是绝对路径
   - 如果指定了 `[config.mtime所在目录]`，验证其必须是绝对路径
   - 确定 `config.mtime` 文件的完整路径：
     - 如果指定了 `[config.mtime所在目录]`，则 `config.mtime` 路径 = `<config.mtime所在目录>/config.mtime`
     - 如果未指定，则 `config.mtime` 路径 = `<当前工作目录>/config.mtime`

2. **配置文件解析**
   - 逐行读取 `config.mtime` 文件
   - 每行格式：`<路径>\t<yyyyMMddHHmmss格式的时间>`
   - 路径可以是相对路径（相对于 `<目录路径>`）或绝对路径
   - 时间格式为 yyyyMMddHHmmss（14位数字），例如：`20231225143045` 表示 2023年12月25日14点30分45秒

3. **文件时间恢复**
   - 如果路径是绝对路径，直接使用该路径
   - 如果路径是相对路径，组合成：`<目录路径>/<相对路径>`
   - 检查文件是否存在
     - 存在：将文件的访问时间（atime）和修改时间（mtime）都设置为记录的时间
     - 不存在：输出警告信息，继续处理下一行，**不会导致程序失败**

4. **跨平台支持**
   - 使用 Go 标准库的 `filepath` 和 `os.Chtimes`，自动适配 Windows 和 Linux 的路径分隔符和时间设置方式

### 完整使用示例

#### Windows 示例

假设你有一个目录 `D:\work\project`，希望记录并恢复其中的文件修改时间：

1. **编译程序：**
```powershell
cd D:\Desktop\mtstamp
go build -o mtstamp.exe
```

2. **记录当前修改时间（config.mtime 保存在当前目录）：**
```powershell
cd D:\Desktop\mtstamp
.\mtstamp.exe log D:\work\project
```
此时会在 `D:\Desktop\mtstamp` 下生成 `config.mtime`。

3. **记录当前修改时间（指定保存目录）：**
```powershell
.\mtstamp.exe log D:\work\project D:\backup
```
此时会在 `D:\backup` 下生成 `config.mtime`。

4. **恢复文件修改时间：**
```powershell
# 不指定config.mtime所在目录，使用当前目录下的config.mtime
.\mtstamp.exe back D:\work\project

# 指定config.mtime所在目录
.\mtstamp.exe back D:\work\project D:\backup
```

#### Linux 示例

假设你有一个目录 `/home/user/project`：

1. **编译程序：**
```bash
cd /home/user/mtstamp
go build -o mtstamp
```

2. **记录当前修改时间：**
```bash
cd /home/user/mtstamp
./mtstamp log /home/user/project
```
此时会在 `/home/user/mtstamp` 下生成 `config.mtime`。

3. **恢复文件修改时间：**
```bash
# 不指定config.mtime所在目录，使用当前目录下的config.mtime
./mtstamp back /home/user/project

# 指定config.mtime所在目录
./mtstamp back /home/user/project /home/user/backup
```

### 注意事项

1. **路径要求**
   - `<目录路径>` 参数**必须使用绝对路径**
   - `[保存目录]` 参数（log命令，如果指定）**必须使用绝对路径**
   - `[config.mtime所在目录]` 参数（back命令，如果指定）**必须使用绝对路径**

2. **配置文件格式**
   - `config.mtime` 采用纯文本格式，每行记录一个文件
   - 格式：`<路径>\t<yyyyMMddHHmmss格式的时间>`
   - 路径使用正斜杠 `/` 分隔，跨平台兼容
   - 时间格式为 yyyyMMddHHmmss（14位数字），例如：`20231225143045` 表示 2023年12月25日14点30分45秒
   - 可以手工查看或编辑
   
   **config.mtime 文件示例：**
   ```
   src/main.go	20231225143045
   src/utils/helper.go	20231225143210
   README.md	20231225143520
   ```

3. **文件变化处理**
   - 如果在生成 `config.mtime` 之后新建了文件，这些新文件不会出现在 `config.mtime` 中，也不会被 `back` 命令修改
   - 如果删除了某些原有文件，`back` 命令会对这些找不到的文件输出警告并跳过，**不会导致程序失败**

4. **跨平台兼容性**
   - 程序使用 Go 标准库，自动适配 Windows 和 Linux 的路径分隔符（`\` 和 `/`）
   - 时间设置使用 `os.Chtimes`，在两个平台都能正常工作

5. **config.mtime 的位置**
   - 使用 `log` 命令时，如果不指定 `[保存目录]`，`config.mtime` 会保存在**运行命令时所在的目录**（当前工作目录），而不是被遍历的目录
   - 使用 `back` 命令时，如果不指定 `[config.mtime所在目录]`，程序会在**运行命令时所在的目录**（当前工作目录）下查找 `config.mtime`
   - 建议明确指定保存/查找位置，避免混淆


