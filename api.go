package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	json "github.com/valyala/fastjson"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	url2 "net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)
//定义全局数据目录
var FILE string="C:\\Users\\hua'wei\\Desktop\\"
type UserInfo struct {
	LoginCode int
	Pwd       string
	Cookie    string
	User      string
	Name      string
	Url       string
}
type Class struct {
	Jxb_id    string
	Jxbmc     string
	Kzmc      string
	Kch       string
	Bklx_id   string
	Do_jxb_id string
	Jxdd      string
	Syxs      string
	Kcmc      string
	Kkxymc    string
	Kcxzmc    string
	Sksj      string
	Jsxx      string
	Jxbrl     string
	Yxzrs     string
	Kklxdm    string
	Kch_id    string
	Year      string
}

func readFile(path string) string {
	c, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(c)
}
func readFileLine(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		os.Create(path)
	}
	file, err = os.Open(path)
	defer file.Close()

	buf := bufio.NewReader(file)
	var result []string
	for {
		line, _, err := buf.ReadLine()

		if err != nil {
			if err == io.EOF { //读取结束，会报EOF
				return result
			}
			return nil
		}
		result = append(result, string(line))
	}
	return result
}
func writeFile(path string, data string) bool {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return false
	}
	defer file.Close()
	buf := bufio.NewWriter(file)
	_, err = buf.Write([]byte(data))
	buf.Flush()
	if err != nil {

		return false
	}

	return true
}
func add(a string, b string) string {
	return a + "\n" + b
}

//正方系统时间点推理算法
func timeTui(fc chan string, userinfo UserInfo) {

	//条件检测
	headdata, _ := getHeadDataXK(userinfo)
	syxs, _ := strconv.Atoi(headdata["syxs"])
	if syxs <= 1 {
		fc <- "[TimeT]:距离开始时间小于1无法完成推理"
		return
	}
	psyxs := syxs
	//推理过程
	for {
		headdata, _ = getHeadDataXK(userinfo)
		syxs, _ := strconv.Atoi(headdata["syxs"])

		if syxs != psyxs {
			t := time.Now()
			timeU := t.Unix()
			timeTemplate := "2006-01-02 15:04:05"
			tm := time.Unix(int64(timeU), 0)
			timeStr := tm.Format(timeTemplate)
			m, _ := time.ParseDuration(strconv.Itoa(psyxs) + "h")
			result := t.Add(m)
			writeFile(FILE+"time.txt", strconv.FormatInt(result.Unix(), 10))
			fc <- add("[TimeT]:定点时间为:"+timeStr, "[TimeT]:推理时间为:"+result.String())
			return
		}
		time.Sleep(1 * time.Minute)
	}
}

func qk(fc chan string, userinfo UserInfo, cstra []string) {
	//构建信息输出器
	var msga string
	msga = add(msga, "--------------------")
	msga = add(msga, "登陆成功:"+userinfo.User+" 姓名:"+userinfo.Name+" url:"+userinfo.Url)

	//获取课程
	headdata, _ := getHeadDataXK(userinfo)
	ca, _ := getClassesXK(userinfo, 2022, "", headdata)

	msga = add(msga, "找到以下教学班:")

	//定义带抢教学数组
	var sca []Class
	for _, v := range cstra {
		for _, v1 := range ca {
			if strings.Index(v1.Jxbmc, v) != -1 {
				msga = add(msga, "名称:"+v1.Jxbmc+" 地点:"+v1.Jxdd)
				sca = append(sca, v1)
			}
		}
	}

	fc <- msga

	//抢课算法
	//时间监听算法，一个小时以外每半小时监听一次，当可监听时间小于监听时间时不断压缩监听时间，当距离抢课开始一分钟时启用秒级监听，当倒数5秒时启用高并发算法
	fc <- "开始抢课....."
	ts := readFile(FILE+"time.txt")
	ut, err := strconv.ParseInt(ts, 10, 64)

	if ut == 0 {
	}
	if err != nil {
		fc <- "[Time]时间解析异常"
		return
	}

	var sleeptime int64
	for {
		t := time.Now()
		tu := t.Unix()
		//日常都是半小时监听频率
		if ut-tu > 3600 {
			fmt.Println("系统正常")
			sleeptime = 1000000000 * 60 * 30
		}
		//两分钟时候更改刷新频率
		if ut-tu <= 120 {
			fmt.Println("刷新频率改变:2mim")
			sleeptime = 1000000000 * 30
		}
		//倒数60s时候更改刷新频率
		if ut-tu <= 60 {
			fmt.Println("刷新频率改变:1s")
			sleeptime = 1000000000
		}
		//倒数20s时候刷新数据
		if ut-tu <= 20 {
			userinfo, err = login(userinfo.Url, userinfo.User, userinfo.Pwd)
			headdata, err = getHeadDataXK(userinfo)
			ca, err = getClassesXK(userinfo, 2022, "", headdata)
		}
		//倒数5s时候进入并发模式
		if ut-tu <= 5 {
			fmt.Println(ut - tu)
		}
		for _, v := range sca {
			res, err := selectClassXK(userinfo, v, headdata)
			if res == "" {
				userinfo, err = login(userinfo.Url, userinfo.User, userinfo.Pwd)
			}
			if err != nil {
				fc <- err.Error()
			}
			fc <- userinfo.Name + " " + v.Jxbmc + " " + res
		}
		///
		time.Sleep(time.Duration(sleeptime))
	}
}

func main() {




	s := readFileLine(FILE+"a.txt")
	fc := make(chan string)

	for _, v := range s {
		va := strings.Split(v, " ")
		user := va[0]
		pwd := va[1]
		userinfo, err := login("http://218.197.80.13/jwglxt", user, pwd)
		userinfo.Name = va[3]
		if userinfo.LoginCode != 200 {
			fmt.Println("登录失败:", userinfo.User, err)
		}

		//携程
		ts := readFile(FILE+"time.txt")
		ut, _ := strconv.Atoi(ts)
		if ut == 0 {
			fmt.Println("启动时间推理......")
			go timeTui(fc, userinfo)
		} else {
			tm := time.Unix(int64(ut), 0)
			timeTemplate := "2006-01-02 15:04:05"
			timeStr := tm.Format(timeTemplate)
			fmt.Println("[time]:解析到时间点:" + timeStr)
		}
		go qk(fc, userinfo, strings.Split(va[2], ","))
	}

	for {
		fmsga := <-fc
		fmt.Println(fmsga)
	}
}

func httpGet(url string, headers map[string]string) (string, map[string][]string, error) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("err:", err)
		return "", nil, err
	}
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		log.Println("err:", err)
		return "", nil, err
	}
	byteds, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("err:", err)
		return "", nil, err
	}
	res.Header.Add("status", res.Status)
	return string(byteds), res.Header, err
}
func httpPost(url string, data string, headers map[string]string) (string, map[string][]string, error) {
	cilent := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Println("err:", err)
		return "", nil, err
	}
	if data != "" {
		req, err = http.NewRequest("POST", url, strings.NewReader(data))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if err != nil {
			log.Println("err:", err)
			return "", nil, err
		}
	}
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	res, _ := cilent.Do(req)
	defer res.Body.Close()

	if err != nil {
		log.Println("err:", err)
		return "", nil, err
	}
	byteds, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("err:", err)
		return "", nil, err
	}

	return string(byteds), res.Header, err
}
func login(url string, user string, pwd string) (UserInfo, error) {
	res, gethh, _ := httpGet(url+"/xtgl/login_getPublicKey.html?time=1662376193387&_=1662376193304", nil)
	cookies := gethh["Set-Cookie"][0]

	var jr json.Parser
	jp, _ := jr.Parse(res)
	Modulus := string(jp.GetStringBytes("modulus"))
	Exponent := string(jp.GetStringBytes("exponent"))
	pubN, err := parse2bigInt(Modulus)
	pubE, err := parse2bigInt(Exponent)

	pub := &rsa.PublicKey{
		N: pubN,
		E: int(pubE.Int64()),
	}

	edata, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(pwd))
	mm := base64.StdEncoding.EncodeToString(edata)

	var hd map[string]string
	hd = make(map[string]string)
	hd["Cookie"] = cookies
	mm = url2.QueryEscape(mm)
	rres, ch, err := httpPost(url+"/xtgl/login_slogin.html?time=1662386627257", "csrftoken=&yhm="+user+"&language=zh_CN&mm="+mm, hd)

	var uinfo UserInfo
	uinfo.Url = url
	uinfo.Pwd = pwd
	uinfo.User = user
	uinfo.LoginCode = 100
	if rres == "" {
		cookies = ch["Set-Cookie"][1]
		uinfo.Cookie = cookies
		uinfo.LoginCode = 200
		return uinfo, err
	}
	return uinfo, err
}
func getHeadDataXK(userinfo UserInfo) (map[string]string, error) {
	var hd map[string]string
	hd = make(map[string]string)
	hd["Cookie"] = userinfo.Cookie
	var r map[string]string
	r = make(map[string]string)
	res, _, err := httpGet(userinfo.Url+"/xsxk/zzxkyzb_cxZzxkYzbIndex.html?gnmkdm=N253512&layout=default&su="+userinfo.User, hd)
	myRegex, err := regexp.Compile("<input type=\"hidden\" name=\"(.*?)\" id=\"(.*?)\" value=\"(.*?)\"/>")
	found := myRegex.FindAllStringSubmatch(res, -1)
	for _, v := range found {
		r[v[1]] = v[3]
	}
	res1, _, err := httpPost(userinfo.Url+"/xsxk/zzxkyzb_cxZzxkYzbDisplay.html?gnmkdm=N253512&su="+userinfo.User, "xkkz_id="+r["firstXkkzId"]+"&xszxzt=1&kspage=0&jspage=0", hd)
	found1 := myRegex.FindAllStringSubmatch(res1, -1)
	for _, v := range found1 {
		r[v[1]] = v[3]
	}
	return r, err
}
func getClassesXK(userinfo UserInfo, year int, term string, headdata map[string]string) ([]Class, error) {
	var hd map[string]string
	hd = make(map[string]string)
	hd["Cookie"] = userinfo.Cookie
	rwlx := headdata["rwlx"]
	bklx_id := headdata["bklx_id"]
	jg_id := headdata["jg_id_1"]
	njdm_id_1 := headdata["njdm_id_1"]
	zyh_id_1 := headdata["zyh_id_1"]
	zyfx_id := headdata["zyfx_id"]
	njdm_id := headdata["njdm_id"]
	bh_id := headdata["bh_id"]
	xslbdm := headdata["xslbdm"]
	ccdm := headdata["ccdm"]
	xsbj := headdata["xsbj"]
	xkxnm := headdata["xkxnm"]
	xkxqm := headdata["xkxqm"]
	kklxdm := headdata["firstKklxdm"]
	res, _, err := httpPost(userinfo.Url+"/xsxk/zzxkyzb_cxZzxkYzbPartDisplay.html?gnmkdm=N253512&su="+userinfo.User, "rwlx="+rwlx+"&xkly=0&bklx_id="+bklx_id+"&sfkkjyxdxnxq=0&xqh_id=1&jg_id="+jg_id+"&njdm_id_1="+njdm_id_1+"&zyh_id_1="+zyh_id_1+"&zyh_id="+zyh_id_1+"&zyfx_id="+zyfx_id+"&njdm_id="+njdm_id+"&bh_id="+bh_id+"&xbm=1&xslbdm="+xslbdm+"&ccdm="+ccdm+"&xsbj="+xsbj+"&sfkknj=0&sfkkzy=0&kzybkxy=0&sfznkx=0&zdkxms=0&sfkxq=0&sfkcfx=0&kkbk=0&kkbkdj=0&sfkgbcx=0&sfrxtgkcxd=0&tykczgxdcs=0&xkxnm="+xkxnm+"&xkxqm="+xkxqm+"&kklxdm="+kklxdm+"&bbhzxjxb=0&rlkz=0&xkzgbj=0&kspage=1&jspage=10&jxbzb=", hd)

	var carr []Class

	var jr json.Parser
	jp, _ := jr.Parse(res)
	ja := jp.GetArray("tmpList")
	for _, v := range ja {

		var c Class
		c.Yxzrs = string(v.GetStringBytes("yxzrs"))
		c.Jxb_id = string(v.GetStringBytes("jxb_id"))
		c.Jxbmc = string(v.GetStringBytes("jxbmc"))
		c.Kch = string(v.GetStringBytes("kch"))
		c.Kch_id = string(v.GetStringBytes("kch_id"))
		c.Kklxdm = string(v.GetStringBytes("kklxdm"))
		c.Kzmc = string(v.GetStringBytes("kzmc"))
		c.Kcmc = string(v.GetStringBytes("kcmc"))
		c.Year = string(v.GetStringBytes("year"))

		carr = append(carr, c)
	}

	kch_id := carr[0].Kch_id
	xkkz_id := headdata["firstXkkzId"]
	//fmt.Println("rwlx=" + rwlx + "&xkly=0&bklx_id=" + bklx_id + "&sfkkjyxdxnxq=0&xqh_id=1&jg_id=" + jg_id + "&njdm_id_1=" + njdm_id_1 + "&zyh_id_1=" + zyh_id_1 + "&zyh_id=" + zyh_id_1 + "&zyfx_id=" + zyfx_id + "&njdm_id=" + njdm_id + "&bh_id=" + bh_id + "&xbm=1&xslbdm=" + xslbdm + "&ccdm=" + ccdm + "&xsbj=" + xsbj + "&sfkknj=0&sfkkzy=0&kzybkxy=0&sfznkx=0&zdkxms=0&sfkxq=0&sfkcfx=0&kkbk=0&kkbkdj=0&sfkgbcx=0&sfrxtgkcxd=0&tykczgxdcs=0&xkxnm=" + xkxnm + "&xkxqm=" + xkxqm + "&kklxdm=" + kklxdm + "&bbhzxjxb=0&rlkz=0&xkzgbj=0&kspage=1&jspage=10&jxbzb=")
	res, _, err = httpPost(userinfo.Url+"/xsxk/zzxkyzbjk_cxJxbWithKchZzxkYzb.html?gnmkdm=N253512&su="+userinfo.User, "rwlx="+rwlx+"&xkly=0&bklx_id="+bklx_id+"&sfkkjyxdxnxq=0&xqh_id=1&jg_id="+jg_id+"&njdm_id_1="+njdm_id_1+"&zyh_id_1="+zyh_id_1+"&zyh_id="+zyh_id_1+"&zyfx_id="+zyfx_id+"&njdm_id="+njdm_id+"&bh_id="+bh_id+"&xbm=1&xslbdm="+xslbdm+"&ccdm="+ccdm+"&xsbj="+xsbj+"&sfkknj=0&sfkkzy=0&kzybkxy=0&sfznkx=0&zdkxms=0&sfkxq=0&sfkcfx=0&kkbk=0&kkbkdj=0&sfkgbcx=0&sfrxtgkcxd=0&tykczgxdcs=0&xkxnm="+xkxnm+"&xkxqm="+xkxqm+"&kklxdm="+kklxdm+"&kch_id="+kch_id+"&jxbzcxskg=0&xkkz_id="+xkkz_id+"&cxbj=0&fxbj=0", hd)

	jp, err = jr.Parse(res)
	ja, err = jp.Array()

	i := 0
	for _, v := range ja {
		carr[i].Do_jxb_id = string(v.GetStringBytes("do_jxb_id"))
		carr[i].Bklx_id = headdata["bklx_id"]
		carr[i].Kcxzmc = string(v.GetStringBytes("kcxzmc"))
		carr[i].Jxbrl = string(v.GetStringBytes("jxbrl"))
		carr[i].Jxdd = string(v.GetStringBytes("jxdd"))
		carr[i].Kkxymc = string(v.GetStringBytes("kkxymc"))
		carr[i].Syxs = headdata["syxs"]
		carr[i].Sksj = string(v.GetStringBytes("sksj"))
		carr[i].Jsxx = string(v.GetStringBytes("jsxx"))
		i++
	}
	return carr, err
	/*
			rwlx=3&xkly=0&bklx_id=714081E2BFA22A10E0530B50C5DA02F6&sfkkjyxdxnxq=0&xqh_id=1&jg_id=300800&zyh_id=1506&zyfx_id=wfx&njdm_id=2022&bh_id=15062201&xbm=1&xslbdm=421&bbhzxjxb=0&ccdm=3&xsbj=4294967296&sfkknj=0&sfkkzy=0&kzybkxy=0&sfznkx=0&zdkxms=0&sfkxq=0&sfkcfx=0&kkbk=0&kkbkdj=0&xkxnm=2022&xkxqm=3&xkxskcgskg=0&rlkz=0&kklxdm=06
		    rwlx=3&xkly=0&bklx_id=714081E2BFA22A10E0530B50C5DA02F6&sfkkjyxdxnxq=0&xqh_id=1&jg_id=300800&njdm_id_1=2022&zyh_id_1=1506&zyh_id=1506&zyfx_id=wfx&njdm_id=2022&bh_id=15062201&xbm=1&xslbdm=421&ccdm=3&xsbj=4294967296&sfkknj=0&sfkkzy=0&kzybkxy=0&sfznkx=0&zdkxms=0&sfkxq=0&sfkcfx=0&kkbk=0&kkbkdj=0&sfkgbcx=0&sfrxtgkcxd=0&tykczgxdcs=0&xkxnm=2022&xkxqm=3&kklxdm=06&bbhzxjxb=0&rlkz=0&xkzgbj=0&kspage=1&jspage=10&jxbzb=

				blyxrs: "0"
				blzyl: "0"
				cxbj: "0"
				date: "二○二二年九月七日"
				dateDigit: "2022年9月7日"
				dateDigitSeparator: "2022-9-7"
				day: "7"
				fxbj: "0"
				jgpxzd: "1"
				jxb_id: "DF53E483D53D763BE0530B50C5DA9400"
				jxbmc: "(2022-2023-1)-TB5902-1毽球"
				jxbzls: "1"
				kch: "TB5902"
				kch_id: "70C7D2FE86793035E0530B50C5DA18FB"
				kcmc: "大学体育(1)"
				kcrow: "1"
				kklxdm: "06"
				kzmc: "体育课"
				listnav: "false"
				localeKey: "zh_CN"
				month: "9"


					jr, _ := json1.NewJson([]byte(res))
					js, _ := jr.Get("tmpList").Array()
					for _, v := range js {
						fmt.Println(v.(map[string]interface{})["date"])
					}*/

}
func selectClassXK(userinfo UserInfo, c Class, headdata map[string]string) (string, error) {
	var hd map[string]string
	hd = make(map[string]string)
	hd["Cookie"] = userinfo.Cookie
	jxb_ids := c.Do_jxb_id
	kch_id := c.Kch_id
	qz := "0"
	xkxnm := c.Year
	xkxqm := headdata["xkxqm"]
	njdm_id := userinfo.User[0:2]
	zyh_id := userinfo.User[2:6]
	kklxdm := c.Kklxdm
	res, _, err := httpPost(userinfo.Url+"/xsxk/zzxkyzb_xkBcZyZzxkYzb.html?gnmkdm=N253512&su="+userinfo.User, "jxb_ids="+jxb_ids+"&kch_id="+kch_id+"&qz="+qz+"&xkxnm="+xkxnm+"&xkxqm="+xkxqm+"&njdm_id="+njdm_id+"&zyh_id="+zyh_id+"&kklxdm="+kklxdm, hd)

	return res, err
}
func getChoosedClassesXK(userinfo UserInfo, headdata map[string]string) ([]Class, error) {
	var hd map[string]string
	hd = make(map[string]string)
	hd["Cookie"] = userinfo.Cookie
	xkly := headdata["xkly"]
	xz := headdata["xz"]
	jg_id := headdata["jg_id_1"]
	zyh_id := headdata["zyh_id"]
	zyfx_id := headdata["zyfx_id"]
	njdm_id := headdata["njdm_id"]
	bh_id := headdata["bh_id"]
	ccdm := headdata["ccdm"]
	xkxnm := headdata["xkxnm"]
	xkxqm := headdata["xkxqm"]
	res, _, err := httpPost(userinfo.Url+"/xsxk/zzxkyzb_cxZzxkYzbChoosedDisplay.html?gnmkdm=N253512&su="+userinfo.User, "jg_id="+jg_id+"&zyh_id="+zyh_id+"&njdm_id="+njdm_id+"&zyfx_id="+zyfx_id+"&bh_id="+bh_id+"&xz="+xz+"&ccdm="+ccdm+"&xkxnm="+xkxnm+"&xkxqm="+xkxqm+"&xkly="+xkly, hd)
	var carr []Class

	//fmt.Println("jg_id=" + jg_id + "&zyh_id=" + zyh_id + "&njdm_id=" + njdm_id + "&zyfx_id=" + zyfx_id + "&bh_id=" + bh_id + "&xz=" + xz + "&ccdm=" + ccdm + "&xkxnm=" + xkxnm + "&xkxqm=" + xkxqm + "&xkly=" + xkly)
	var jr json.Parser
	jp, err := jr.Parse(res)
	ja, err := jp.Array()
	for _, v := range ja {

		var c Class
		c.Yxzrs = string(v.GetStringBytes("yxzrs"))
		c.Jxb_id = string(v.GetStringBytes("jxb_id"))
		c.Jxbmc = string(v.GetStringBytes("jxbmc"))
		c.Kch = string(v.GetStringBytes("kch"))
		c.Kch_id = string(v.GetStringBytes("kch_id"))
		c.Kklxdm = string(v.GetStringBytes("kklxdm"))
		c.Kcmc = string(v.GetStringBytes("kcmc"))
		c.Year = string(v.GetStringBytes("year"))
		c.Do_jxb_id = string(v.GetStringBytes("do_jxb_id"))
		c.Bklx_id = headdata["bklx_id"]
		c.Kcxzmc = string(v.GetStringBytes("kcxzmc"))
		c.Jxbrl = string(v.GetStringBytes("jxbrs"))
		c.Jxdd = string(v.GetStringBytes("jxdd"))

		c.Syxs = headdata["syxs"]
		c.Sksj = string(v.GetStringBytes("sksj"))
		c.Jsxx = string(v.GetStringBytes("jsxx"))

		carr = append(carr, c)
	}
	return carr, err

}
func parse2bigInt(s string) (bi *big.Int, err error) {
	bi = &big.Int{}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return
	}
	bi.SetBytes(b)
	return
}
func mapToStr(m *map[string]string) string {
	r := ""
	for k, v := range *m {
		r += "&" + k + "=" + v
	}
	return r[1:]
}
