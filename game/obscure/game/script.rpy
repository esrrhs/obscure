# 游戏的脚本可置于此文件中。

# 声明此游戏使用的角色。颜色参数可使角色姓名着色。

define me = Character("我")

define suan = Character("算命先生")

define pm = Character("PM")

define luren = Character("路人")

define e = Character("艾琳")


# 游戏在此开始。

label start:

    # 显示一个背景。此处默认显示占位图，但您也可以在图片目录添加一个文件
    # （命名为 bg room.png 或 bg room.jpg）来显示。

    scene bg 000100

    me "周一周一，速速归西"
    me "殆哉！殆哉！"
    "如故，余且怨，骑乘而上班"
    me "吾草，又雨。"

    show suan normal at right

    suan "公且慢，汝今必有大灾！"

    me "哦？何灾？"

    suan "天机不可泄露"

    me "汝将无母也"

    "遂不理，去公司上班"

    scene bg 000200

    me "此乐问之乐，不思蜀也"

    me "\"吾有万贯家财，想购置京城房屋？\"又一装逼小儿，可恨！"

    show pm normal at right

    pm "君TAPD单遗留甚多，如之奈何？"

    me "莫急，待吾加班数日，即可"

    pm "甚好"

    hide pm

    me "再刷乐问两时辰先"

    "."

    ".."

    "..."

    me "好浓之烟味，不好！"

    scene bg 000300

    "才下雨，竟着火"

    "今果有大灾"

    scene bg 000200

    me "不过火势不大，待洒家再写两行代码，转完TAPD，再撤不迟"

    "..."

    ".."

    "."

    me "不好，网已断，吾命休矣"

    scene bg 000400

    "..."

    ".."

    "."

    me "我擦，我怎么还活着，这里是哪里？"

    me "等一等，我怎么说话也变了"

    show luren normal

    luren "小伙子，刚穿越来的吧？"

    me "又是穿越？吐了"

    luren "严格意义也不是穿越，你们的世界死掉的人就会到我们这个世界"

    luren "就像过滤装置那样，一层一层的"

    luren "你现在这地方，就是传送点，今天来了一大波人"

    me "哦，那最后一层是哪里？"

    luren "轮回，就像蛇咬住自己的尾巴"

    hide luren

    "消失了？"

    "这一看就是NPC，搁这新手引导呢"

    "不过"

    "这世界看起来不错，画风可以"

    "肚子有点饿，去找点吃的"

    scene bg 000500

    "迷路了"

    "不知不觉，走到这森林里"

    "没办法，只能硬着头皮继续往前走"

    scene bg 000501

    "周一周一，马上归西"

    scene bg 000600

    "周一周一，马上归西"

    scene bg 000601

    "周一周一，马上归西"

    scene bg 000700

    "周一周一，马上归西"

    scene bg 000800

    "周一周一，马上归西"

    scene bg 000900

    "周一周一，马上归西"

    scene bg 001000

    "周一周一，马上归西"

    scene bg 001100

    "周一周一，马上归西"

    scene bg 001200

    "周一周一，马上归西"

    scene bg 001300

    "周一周一，马上归西"

    scene bg 001400

    "周一周一，马上归西"

    scene bg 001500

    # 显示角色立绘。此处使用了占位图，但您也可以在图片目录添加命名为
    # eileen happy.png 的文件来将其替换掉。

    show eileen happy

    # 此处显示各行对话。

    e "您已创建一个新的 Ren'Py 游戏。"

    e "当您完善了故事、图片和音乐之后，您就可以向全世界发布了！"

    # 此处为游戏结尾。

    return
