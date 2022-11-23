import subprocess
import threading

from flask import Flask, render_template, request, redirect, url_for, session  # 导入render_template模块
import os
import redis

# from transaction import index1         #导入news蓝图
# from index-2 import  index-2        #导入user蓝图
# from product import product   #导入product蓝图
app=Flask(__name__)
# app.secret_key='any random string'

# urls=[index1]      #将三个路由构建数组
# for url in urls:
#     app.register_blueprint(url)   #将三个路由均实现蓝图注册到主app应用上

red_obj = redis.Redis(host = "127.0.0.1", port = 6379, db=0)
red_obj_10 = redis.Redis(host = "127.0.0.1", port = 6379, db=10)
red_obj_11 = redis.Redis(host = "127.0.0.1", port = 6379, db=11)

def exemain1():
    # path1 = r"D:\Nodes\start.bat"
    path1 = "start.bat"
    print(f'"{path1}"')
    print(os.path.exists(path1))
    # subprocess.Popen(path1)
    # return os.system(path1)
    os.startfile(path1)
    # os.popen(path1)
def exemain2():
    path1 = "start2.bat"
    print(f'"{path1}"')
    print(os.path.exists(path1))
    os.startfile(path1)
def exemain3():
    path1 = "start3.bat"
    print(f'"{path1}"')
    print(os.path.exists(path1))
    os.startfile(path1)
def exemain4():
    path1 = "start4.bat"
    print(f'"{path1}"')
    print(os.path.exists(path1))
    os.startfile(path1)
def exemain5():
    path1 = "start5.bat"
    print(f'"{path1}"')
    print(os.path.exists(path1))
    os.startfile(path1)
def exemain6():
    path1 = "start6.bat"
    print(f'"{path1}"')
    print(os.path.exists(path1))
    os.startfile(path1)
def exemain7():
    path1 = "start7.bat"
    print(f'"{path1}"')
    print(os.path.exists(path1))
    os.startfile(path1)
def exemain8():
    path1 = "start8.bat"
    print(f'"{path1}"')
    print(os.path.exists(path1))
    os.startfile(path1)
def exemain9():
    path1 = "start9.bat"
    print(f'"{path1}"')
    print(os.path.exists(path1))
    os.startfile(path1)
def exemain10():
    path1 = "start10.bat"
    print(f'"{path1}"')
    print(os.path.exists(path1))
    os.startfile(path1)

@app.route('/')
def index():
    # msg = ""
    return render_template("index.html")

@app.route('/getnode')
def getnodepage():
    # msg = ""
    return render_template("getnode.html")

@app.route('/transactionsearch')
def transactionsearch():
    # msg = ""
    return render_template("BlockDisplay.html")

@app.route('/index-5')
def index5page():
    # msg = ""
    return render_template("index-5.html")
@app.route('/index-6')
def index6page():
    # msg = ""
    return render_template("index-6.html")

@app.route('/MineStarting', methods=['POST','GET'])
def MineStarting():
    result1 = "111"
    if request.method == 'POST':
        # t1 = threading.Thread(target=exeserver1, args=())
        # t2 = threading.Thread(target=exeserver2, args=())
        # t3 = threading.Thread(target=exeserver3, args=())
        # t4 = threading.Thread(target=exeserver4, args=())
        # t5 = threading.Thread(target=exeserver5, args=())
        # t6 = threading.Thread(target=exeserver6, args=())
        # t7 = threading.Thread(target=exeserver7, args=())
        # t8 = threading.Thread(target=exeserver8, args=())
        # t9 = threading.Thread(target=exeserver9, args=())
        # t10 = threading.Thread(target=exeserver10, args=())
        # t11 = threading.Thread(target=exemain1, args=())
        # t12 = threading.Thread(target=exemain2, args=())
        # t13 = threading.Thread(target=exemain3, args=())
        # t14 = threading.Thread(target=exemain4, args=())
        # t15 = threading.Thread(target=exemain5, args=())
        # t16 = threading.Thread(target=exemain6, args=())
        # t17 = threading.Thread(target=exemain7, args=())
        # t18 = threading.Thread(target=exemain8, args=())
        # t19 = threading.Thread(target=exemain9, args=())
        # t20 = threading.Thread(target=exemain10, args=())
        # for i in range(1,11):
        #     _t = "t"+str(i)
        #     eval(_t).start()
        #     # eval(_t).join()
        # for i in range(1,11):
        #     _t = "t"+str(i)
        #     eval(_t).join()

        count = 0
        for i in range(1,11):

            i_tmp = 8000+i
            node_num = "0.0.0.0:"+str(i_tmp)
            # bool_num = red_obj.get(node_num)
            # print(bool_num)
            # s = bytes.decode(bool_num)
            # if s == "T" :
            order = "'"+"netstat -ano|findstr "+node_num+"'"
            print(order)
            num_status = os.system(eval(order))
            # print(i_tmp +"的端口状态"+ num_status)
            if num_status == 0:
                _m = "exemain" + str(i)
                n = threading.Thread(target=eval(_m), args=())
                n.start()
                n.join()
                # _t = "t" + str(i+10)
                # eval(_t).start()
                # eval(_t).join()
                count = count + 1
        result1 = "Done"
        print(count)
    return render_template("index-6.html", data = result1)

# def exeserver1():
#     path = ".\Miner" + str(1) + "\myserver\myserver.exe"
#     print(path)
#     res = os.system(path)
# def exeserver2():
#     path = ".\Miner" + str(2) + "\myserver\myserver.exe"
#     print(path)
#     res = os.system(path)
# def exeserver3():
#     path = ".\Miner" + str(3) + "\myserver\myserver.exe"
#     print(path)
#     res = os.system(path)
# def exeserver4():
#     path = ".\Miner" + str(4) + "\myserver\myserver.exe"
#     print(path)
#     res = os.system(path)
# def exeserver5():
#     path = ".\Miner" + str(5) + "\myserver\myserver.exe"
#     print(path)
#     res = os.system(path)
# def exeserver6():
#     path = ".\Miner" + str(6) + "\myserver\myserver.exe"
#     print(path)
#     res = os.system(path)
# def exeserver7():
#     path = ".\Miner" + str(7) + "\myserver\myserver.exe"
#     print(path)
#     res = os.system(path)
# def exeserver8():
#     path = ".\Miner" + str(8) + "\myserver\myserver.exe"
#     print(path)
#     res = os.system(path)
# def exeserver9():
#     path = ".\Miner" + str(9) + "\myserver\myserver.exe"
#     print(path)
#     res = os.system(path)
# def exeserver10():
#     path = ".\Miner" + str(10) + "\myserver\myserver.exe"
#     print(path)
#     res = os.system(path)



@app.route('/transactiondisplay')
def index1page():
    return render_template("balancedisplay.html")

@app.route('/processDisplay')
def processDisplay():
    return render_template("processDisplay.html")

@app.route('/lasttwobits', methods=['POST','GET'])
def LasttwoBits():
    if request.method == 'POST':
        lastbits = red_obj_10.get('Cake')
        print(lastbits)
    return render_template("processDisplay.html",data = bytes.decode(lastbits))
@app.route('/getwinner', methods=['POST','GET'])
def GetWinner():
    if request.method == 'POST':
        winneradd = red_obj_10.get('Winner')
        print(winneradd)
    return render_template("processDisplay.html",winnerdata = bytes.decode(winneradd))
@app.route('/getallwallet', methods=['POST','GET'])
def GetAllWallet():
    walletlist = []
    if request.method == 'POST':
        count_wallet = 0
        for i in range(1, 11):
            i_tmp1 = 8000 + i
            node_num = "0.0.0.0:" + str(i_tmp1)
            order_wallet = "'" + "netstat -ano|findstr " + node_num + "'"
            print(order_wallet)
            num_status = os.system(eval(order_wallet))
            if num_status == 0:
                count_wallet = count_wallet + 1
        for i in range(0, count_wallet):
            walletinfo = red_obj_10.get(i)
            walletlist.append(bytes.decode(walletinfo))
            print(walletlist)
    return render_template("processDisplay.html",walletlist = walletlist)

@app.route('/POTblocksdisplay', methods=['POST','GET'])
def POTBlocksDisplay():
    blocklist = []
    if request.method == 'POST':
        signal = int(bytes.decode(red_obj.get("index")))
        for i in range(1,signal+1):
            blockinfo = red_obj.get("B"+str(i))
            blocklist.append(bytes.decode(blockinfo))
            print(blocklist)
    return render_template("BlockDisplay.html",blocklist = blocklist)

@app.route('/getbalance', methods=['POST','GET'])
def GetBalance():
    nodelist = ["A","B","C","D","E","F","G","H","I","J"]
    coinAlist = []
    coinBlist = []
    if request.method == 'POST':
        for i in range(0,10):
            coinAinfo = red_obj_11.get(nodelist[i]+str(1))
            coinBinfo = red_obj_11.get(nodelist[i]+str(0))
            coinAlist.append(bytes.decode(coinAinfo))
            coinBlist.append(bytes.decode(coinBinfo))
            print(coinAlist)
            print(coinBlist)
    return render_template("balancedisplay.html",coinAlist = coinAlist, coinBlist = coinBlist)

if __name__=="__main__":
    print(app.url_map)                 #打印url结构图
    app.run(port=3030,host="127.0.0.1",debug=True)











# @app.route('/registernode')
# def registernodepage():
#     # msg = ""
#     return render_template("registernode.html")
#
# @app.route('/registernodeprocess', methods=['POST','GET'])
# def registernodeprocess():
#     # msg = ""
#     if request.method == 'POST':
#         address = request.form['nodeaddress']
#         name = request.form['nodename']
#         if address == '123' and name == '123':
#             result = 'registernode successfully'
#             print(result)
#         else:
#             result = 'registernode unsuccessfully'
#     return render_template("registernode.html",data = result)
#
#
# @app.route('/getnodeprocess', methods=['POST','GET'])
# def getnodeprocesspage111():
#     # msg = ""
#     if request.method == 'POST':
#         address = request.form['nodeaddress']
#         name = request.form['nodename']
#         if address == '123' and name == '123':
#             result = 'createnode successfully'
#             print(result)
#         else:
#             result = 'createnode unsuccessfully'
#     return render_template("getnode.html",data = result)
#
# @app.route('/transactioncreate')
# def index2page():
#     # msg = ""
#     return render_template("transactioncreate.html")
#
# @app.route('/transactioncreateprocess', methods=['POST','GET'])
# def createprocess():
#     # msg = ""
#     if request.method == 'POST':
#         address = request.form['add']
#         name = request.form['name']
#         if address == '123' and name == '123':
#             result = 'Tx create successfully'
#             print(result)
#         else:
#             result = 'Tx create unsuccessfully'
#     return render_template("transactioncreate.html",data = result)
#
# @app.route('/newstest',methods=['POST','GET'])
# def newstestpage():
#     if request.method=='POST':
#         nm = request.form['add']
#         print(nm)
#         if nm == '123':
#             newsli = 323
#         elif nm == '321':
#             newsli = 'rajcnsja'
#
#     return render_template("news.html",data=newsli)
# @app.route('/POTsearchprocess', methods=['POST','GET'])
# def transactionsearchprocess():
#     # msg = ""
#     if request.method == 'POST':
#         address = request.form['add']
#         name = request.form['name']
#         if address == '123' and name == '123':
#             result = 'Tx search successfully'
#             print(result)
#         else:
#             result = 'Tx search unsuccessfully'
#     return render_template("BlockDisplay.html",data = result)
#
# @app.route('/chaindisplay')
# def chaindisplaypage():
#     # msg = ""
#     return render_template("chaindisplay.html")
#
# @app.route('/chaindisplayprocess', methods=['POST','GET'])
# def chaindisplayprocess():
#     # msg = ""
#     if request.method == 'POST':
#         address = request.form['add']
#         name = request.form['name']
#         if address == '123' and name == '123':
#             result = 'chain search successfully'
#             print(result)
#         else:
#             result = 'chain search unsuccessfully'
#     return render_template("chaindisplay.html",data = result)


    # num_result = red_obj.sadd("Cake","333")
    # num_result1 = red_obj.sadd("Winner","correctaddress")
    # num_result = red_obj.set("Cake","333")
    # num_result1 = red_obj.set("Winner","correctaddress")
    # testwallets1 = red_obj.set("101", "correctaddress11,corecssj11,balance1:")
    # testwallets2 = red_obj.set("102", "correctaddress22,corecssj1,balance1:")
    # testwallets = red_obj.set("103", "correctaddress33,corecssj1,balance1:")
    # blockinfo1 = red_obj.set("B1", "correctaddress11,corecssj11,balance1:")
    # blockinfo2 = red_obj.set("B2", "correctaddress22,corecssj1,balance1:")
    # blockinfo3 = red_obj.set("B3", "correctaddress33,corecssj1,balance1:")
    # cinfo1 = red_obj.set("X1", "balanceX1:")
    # cinfo2 = red_obj.set("Y1", "balanceY1:")
    # cinfo3 = red_obj.set("Z1", "balanceZ1:")
    # result = red_obj.set('name','skylark')
    # result = red_obj.get('Cake')
    # print(num_result)
    # print(num_result1)
    # print(result)
    # blocklist = []
    # signal = 4
    # for i in range(1,signal):
    #     block = red_obj.get("B"+str(i))
    #     blocklist.append(bblock)
    #     print(blocklist)
    # nodelist = ["X","Y","Z"]
    # testlist = []
    # for i in range(0,3):
    #     testinfo = red_obj.get(nodelist[i]+str(1))
    #     testlist.append(bytes.decode(testinfo))
    #     print(testlist)

    # for i in range(1,2):
    #     path = "./Miner"+str(i)+"/test.txt"
    #     print(path)
    #     with open(path,mode='r',encoding='utf-8') as f:
    #         data = f.readlines()
    #         print(data[0])
    #         print(data[1])
