# 游戏的脚本可置于此文件中。

# 声明此游戏使用的角色。颜色参数可使角色姓名着色。

define me = Character("我")

define suan = Character("算命先生")

define pm = Character("PM")

define luren = Character("路人")

define zhuyao = Character("猪妖")

define laoren = Character("老朱")

define laowang = Character("老朱")

define nvhai = Character("女孩")

define yuefu = Character("岳父")

define e = Character("艾琳")

define sun = Character("孙大圣")

define zhu = Character("猪八戒")

define sha = Character("沙僧")

define tang = Character("唐僧")

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

    scene bg 000301

    "."

    ".."

    "..."

    "这个是死前产生的幻觉吗？"

    "..."

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

    me "!!!"

    me "什么玩意？"

    "走近看看"

    "这地上是血"

    "奇怪的建筑里，散发着一股股臭气"

    "里面摆放了成山的肉"

    me "溜了溜了"

    scene bg 000600

    "啊!!!!!"

    "眼前突然出现一个巨大的黑猪"

    "正冲我张开血盆大口"

    "看似马上要吃了我"

    "我吓在原地，想跑，腿却无法动弹"

    "..."

    zhuyao "小子，你知道擅闯禁地是什么下场吗？"

    "..."

    "猪也会说话？"

    zhuyao "本来呢，我应该吃了你"

    me "..."

    zhuyao "但是你太柴了，不对我胃口"

    zhuyao "而且一看你，就知道是从上层世界来的"

    me "..."

    zhuyao "我不吃你，我要你干个活"

    me "额，什么？"

    zhuyao "去我家，做我的女儿的宠物"

    "..."

    "...."

    "....."

    me "如果我不同意呢"

    zhuyao "那我只能吃了你"

    me "好吧"

    zhuyao "走，跟我回家"

    scene bg 000601

    me "猪先生，还有多远，天都亮了"

    show laoren normal

    laoren "马上到了"

    me "你是？？"

    laoren "我是老朱啊"

    me "..."

    laoren "哦，忘记你不懂了"
    
    laoren "我们妖怪白天是正常人，晚上才变成妖怪"

    me "哦这样啊"

    me "那你女儿岂不是也是？"

    laowang "不是"

    "老朱不再说话，闷头继续往前走"

    scene bg 000700

    laowang "前面的房子就是了"

    "我停下脚步"

    "一边休息，一边欣赏这里的风景"

    me "朱先生，这地方挺漂亮"

    show laowang happy at right

    laowang "那还用说"

    scene bg 000800

    "呼...呼..."

    show laowang normal

    laowang "这就是我家，气派吧"

    laowang "坐这等会，我女儿马上就来"

    "想起之前老朱说给她女儿当宠物"

    "不由得忐忑"

    me "朱先生，你刚才说，你是白天人晚上妖"

    me "但是你女儿不是，那她是什么样的？"

    laowang "急啥"

    laowang "一会你看到不就知道了？"

    hide laowang

    "一阵脚步声传来"

    "越来越近"

    show nvhai happy

    nvhai "你好"

    "想不到眼前的可爱少女，竟然是老朱的女儿"

    show laowang normal at right

    laowang "女儿，这是我给你找来的宠物"

    nvhai "别听他胡说，我爸他看我孤苦伶仃"

    nvhai "特地找你来做我的朋友呢"

    laowang "哈哈，小子，这是我女儿艾琳"

    laowang "她妈和你一样，也是你们那个世界的人"

    laowang "后来遇到了我，就有了艾琳，结果难产死了"

    laowang "艾琳呢，随他妈，也是跟你一样的人"

    me "哦，这样啊"

    e "对啊，放心啦，我不吃你"
    
    e "饿了没有？"

    me "有点饿了"

    e "走吧，我们去吃饭吧"

    scene bg 000900

    me "..."

    me "你们每顿就吃这个吗？"

    e "对啊，这都是我们自己养的，很健康"

    e "怎么了，吃不下吗？"

    me "我想吃熟的"

    e "那我让后厨加工下吧"

    e "太可惜了，你少了好多美味"

    "..."

    scene bg 001000

    "第二天，以及之后的每一天，我在艾琳家住下来了"

    "这里地处偏僻，风景宜人"

    show e happy at left

    e "快看，太阳出来啦！"

    me "嗯，很美"

    scene bg 001100

    "院里种有花园，各种不知名的花"

    "每天都陪艾琳来这里赏花"

    scene bg 001200

    "这里就是养殖的农场"

    show e happy

    e "走，我带你进去快看"

    scene bg 001300

    e "这是我们养的动物，可爱吧"

    me "这是猪还是人？"

    e "你管它呢？好吃就行了"

    me "....."

    scene bg 001400

    show e normal at left

    e "怎么样，这里漂亮吧？"

    e "我每天都喜欢来这里，坐着发呆一整天"

    "于是，我也经常陪着艾琳来这里发呆"

    "这样的平静的日子重复了一天又一天"

    "已经不知道过了多久"

    "终于有一天"

    "一个不速之客到来"

    scene bg 001500

    show sun angry at left

    sun "爷爷我是孙悟空，可认识我吗？"

    show laowang afraid at right

    laowang "认识，认识"

    sun "老子护送唐僧西天取经，路过此地，责任重大"

    sun "方圆五里，不能有一丝的脏乱"

    sun "房屋破的该修补，路不平的就填平"

    sun "懂了吗？"

    laowang "懂了，懂了"

    sun "懂了？你的家在哪，带路！我去检查下有无隐患"

    sun "可不能影响我师父的取经大业"

    laowang "嗯嗯，大圣，这边请"

    scene bg 001600

    show sun angry at left

    sun "你带我来的这什么破地方？"

    show laowang angry at right

    laowang "你的葬身之地！"

    "说完，老朱一棒子将孙大圣打死在地"

    laowang "什么狗屁大圣，敢来我这撒野？"

    "我大惊，孙大圣我是知道的"

    "想不到，竟然这么容易打死"

    scene bg 001700

    "须臾之间，大圣显出了原形"

    laowang "原来只是个死猴子"

    "然后，他俯下身，在尸体上掏弄半天"

    scene bg 001800

    laowang "找到了！这死猴的原神"

    laowang "吃了就可以延年益寿"

    e "啊，老爹给我！"

    "说完，艾琳一口吞了下去"

    "......"

    "果然，吃完后艾琳感觉更有活力了"
    
    scene bg 001900

    "就这样，又过了几天"

    "终于，又有一行人来到"

    show zhu angry at left

    show tang normal

    show sha normal at right

    zhu "喂，我大师兄几天前来你这，看到过吗？"

    laowang "没有"

    zhu "放屁"

    zhu "我亲眼看你带他进了山洞"

    zhu "之后，我大师兄就再也没出来"

    "老朱见事情败露，目露凶光"

    laowang "是又怎么样，他已经被我宰了吃了"

    tang "这位施主，你可知我是谁？"

    laowang "不知"

    tang "我乃东土大唐高僧"

    tang "皇帝是我结拜兄弟，周边各国政要也我的朋友"

    tang "就连天庭，我都有点人脉"

    tang "你可知，得罪我的下场？"

    laowang "不知道"

    laowang "我管你那么多"

    "说完，老朱又一棒子打死唐僧在地"

    hide tang

    "事发突然，竟无一人反应过来"

    "愣了片刻，猪八戒拔腿就跑"

    hide zhu

    zhu "沙师弟，你顶上，我去搬救兵"

    "说完，没影了"

    scene bg 002000

    "老朱瞬间与沙僧斗在了一起"

    "从地上打到了天上"

    "大战数百回合，难解难分"

    scene bg 001900

    "突然，沙僧跳出来，罢手"

    show sha normal

    sha "去他妈的，这经不取也罢"

    sha "我平日忍辱负重，本来就是想图体制内安稳"

    sha "现在上司被你打死，我也懒得再呆了"

    "说罢，转身离去"

    scene bg 002100

    "当晚，唐僧肉被煮成了肉汤"

    show laoren normal

    laoren "小子，你不吃吗？"

    "我摇了摇头"

    laoren "我查了下，那人真是唐僧"

    laoren "我打死了唐僧，马上就会有一大波麻烦找上门来"

    laoren "但是如果吃了唐僧肉真的能得道升仙，那老子也不怕了"

    show e sad at right

    e "对啊，你也吃吧"

    e "不然我一个人在天上，我会想你的"

    me "好，你们先吃"

    me "管用，我也马上吃"

    "老朱和艾琳喝起了汤"

    "突然，两人浑身散发金光，慢慢飘向了天空"

    laoren "哈哈哈，成了"

    e "到你了，快喝吧"

    "我摇了摇头"

    "掏出事先准好的毒药服下"

    scene bg 000301

    "对不起，我想去看看下一层的世界"

    # 此处为游戏结尾。

    return
