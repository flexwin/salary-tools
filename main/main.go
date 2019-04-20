package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx"
	"html/template"
	"log"
	"net"
	"net/smtp"
	"os"
	"regexp"
	"strings"
)

var config Config

func main() {
	parseConfig()
	salaryFilePath := "./salary.xlsx"
	salaries := getSalaryList(salaryFilePath)
	staffFilePath := "./staff.xlsx"
	staffs := getStaffList(staffFilePath)
	sendMail(salaries, staffs)
}

func getStaffList(filePath string) map[string]Staff {
	excel, err := xlsx.OpenFile(filePath)
	if err != nil {
		log.Fatalln("err:", err.Error())
	}
	staffs := make(map[string]Staff)
	for _, sheet := range excel.Sheets {
		curId := ""
		rowId := 0
		for _, row := range sheet.Rows {
			rowId++
			// 表头
			if rowId == 1 && !config.Staff.NoneHeader {
				continue
			}
			cells := getCellValues(row)
			count := 0
			for _, c := range cells {
				if len(c) > 0 {
					count++
				}
			}
			//空行
			if count < 1 {
				continue
			}
			// 至少要包含身份证和邮箱项
			if count < 2 {
				fmt.Printf("第%d行员工信息读取失败(至少要包含身份证和邮箱项),跳过...\n", rowId)
				continue
			}
			if isIdCard, idCard := isIdCardRow(cells); isIdCard {
				curId = idCard
			} else {
				fmt.Printf("第%d行员工信息读取失败(不包含身份证号或者身份证号格式错误),跳过...\n", rowId)
				continue
			}
			// 如果行包含电子邮件，创建一个新字典项
			if isEmail, mail := isEmailRow(cells); isEmail {
				staffs[curId] = Staff{
					//Name:   cells[0],
					IdCard: curId,
					Mail:   mail,
					//Mobile: cells[3],
				}
			} else {
				fmt.Printf("第%d行员工信息读取失败(不包含邮箱地址),跳过...\n", rowId)
			}
		}
	}
	return staffs
}

func getSalaryList(filePath string) map[string][]SalaryTable {
	excel, err := xlsx.OpenFile(filePath)
	if err != nil {
		log.Fatalln("err:", err.Error())
	}
	salaries := make(map[string][]SalaryTable)
	salaryHeaderCells := make([]string, 50)
	headerLen := 0
	for _, sheet := range excel.Sheets {
		curId := ""
		rowId := 0
		for _, row := range sheet.Rows {
			rowId++
			cells := getCellValues(row)
			count := 0
			for _, c := range cells {
				if len(c) > 0 {
					count++
				}
			}

			// 表头，必须配置
			if rowId == 1 /*&& config.Salary.NoneHeader*/ {
				salaryHeaderCells = cells
				headerLen = count
				continue
			}

			//空行
			if count < 1 {
				continue
			}
			// 至少要包含身份证和邮箱项
			/*if count < 27 {
				fmt.Printf("第%d行薪资信息读取失败(缺少必要数据或者不是明细项),跳过...\n", rowId)
				continue
			}*/
			if isIdCard, idCard := isIdCardRow(cells); isIdCard {
				curId = idCard
			} else {
				fmt.Printf("第%d行薪资信息读取失败(不包含身份证号或者身份证号格式错误),跳过...\n", rowId)
				continue
			}
			var tables []SalaryTable
			for i := 0; i < headerLen; i++ {
				tables = append(tables, SalaryTable{
					Header:  salaryHeaderCells[i],
					Content: cells[i],
				})
			}
			fmt.Println(tables)
			salaries[curId] = tables
			/*salaries[curId] = Salary{
				Id:                cells[0],  // 人员编号
				Name:              cells[1],  // 姓名
				IdCard:            cells[2],  // 身份证号
				BasicWage:         cells[3],  // 基本工资
				PostWage:          cells[4],  // 岗位工资
				XBonus:            cells[5],  // 年资津贴
				WageHike:          cells[6],  // 加薪
				RedPackets:        cells[7],  // 开门红
				MeritPay:          cells[8],  // 绩效
				AllowanceSubsidy:  cells[9],  // 防寒暑费
				HolidaySubsidy:    cells[10], // 节日费
				TotalPay:          cells[11], // 应发合计
				BasicPension:      cells[12], // 基本养老金
				MedicareBenefits:  cells[13], // 医疗保险金
				InsuranceBenefits: cells[14], // 失业保险金
				HousingFund:       cells[15], // 住房公积金
				EnterpriseAnnuity: cells[16], // 企业年金
				LaborFee:          cells[17], // 工会费
				HouseRent:         cells[18], // 房租费
				Electricity:       cells[19], // 电费
				PropertyFee:       cells[20], // 物业费
				SpecialDeduction:  cells[21], // 专项扣除项目
				TaxBase:           cells[22], // 扣税基数
				DeductOrRefundTax: cells[23], // 补扣退个税
				CurrentTax:        cells[24], // 本月扣税
				TotalDeduct:       cells[25], // 实际扣款合计
				TotalIncome:       cells[26], // 实发合计
				Signature:         cells[27], // 签名
				BankName:          cells[28], // 银行名称
				BankAccount:       cells[29], // 银行账号
				//Html: fmt.Sprintf("<tr><td colspan='%d'>%s</td></tr>", len(cells), strings.Join(cells, ""))
			}*/
		}
	}
	return salaries
}

func getCellValues(r *xlsx.Row) (cells []string) {
	for _, cell := range r.Cells {
		txt := strings.Replace(strings.Replace(cell.Value, "\n", "", -1), " ", "", -1)
		cells = append(cells, txt)
	}
	return
}

func isEmailRow(r []string) (isEmail bool, mail string) {
	reg := regexp.MustCompile(`^[a-zA-Z_0-9.-]{1,64}@([a-zA-Z0-9-]{1,200}.){1,5}[a-zA-Z]{1,6}$`)
	for _, v := range r {
		if reg.MatchString(v) {
			return true, v
		}
	}
	return false, ""
}

func isIdCardRow(r []string) (isIdCard bool, idCard string) {
	reg := regexp.MustCompile(`(^[1-9]\d{5}(18|19|([23]\d))\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$)|(^[1-9]\d{5}\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}$)`)
	for _, v := range r {
		if reg.MatchString(v) {
			return true, v
		}
	}
	return false, ""
}

func sendMail(salaries map[string][]SalaryTable, staffs map[string]Staff) {
	fmt.Printf("共需要发送%d封邮件!\n", len(salaries))
	index := 1
	for idCard, salary := range salaries {
		staff, exists := staffs[idCard]
		if !exists {
			fmt.Printf("第%d封邮件发送失败!没有配置接收者(%s)邮箱\n", index, idCard)
			continue
		}

		// 模版生成邮件内容
		var tpl bytes.Buffer
		t, _ := template.ParseFiles("templates/salary_mail.html")
		// fmt.Println(t.Name())
		if err := t.Execute(&tpl, struct {
			CompanyName  string
			SalaryPhone  string
			SalaryTables []SalaryTable
		}{CompanyName: config.Company.Name, SalaryPhone: config.Company.SalaryPhone, SalaryTables: salary}); err != nil {
			fmt.Printf("第%d封邮件生成失败!内容(%s)\n", index, salary)
		}
		message := tpl.String()

		fmt.Printf("第%d封邮件发送中...接收者(%s)邮箱(%s)内容(%s)\n", index, idCard, staff.Mail, salary)
		if err := sendToMail(
			staff.Mail,
			message,
			config); err != nil {
			fmt.Printf("第%d封邮件发送失败!接收者(%s)邮箱(%s)错误(%s)\n", index, idCard, staff.Mail, err.Error())
		} else {
			fmt.Printf("第%d封邮件发送成功.接收者(%s)邮箱(%s)\n", index, idCard, staff.Mail)
		}
		index++
	}
}

func sendToMail(to string, body string, config Config) error {
	mailConfig := config.Mail
	auth := smtp.PlainAuth("", mailConfig.EMail, mailConfig.Password, mailConfig.Host)

	header := make(map[string]string)
	header["From"] = mailConfig.Name + " <" + mailConfig.EMail + ">"
	header["To"] = to
	header["Subject"] = "工资条"
	header["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	err := SendMailUsingTLS(
		fmt.Sprintf("%s:%d", mailConfig.Host, mailConfig.Port),
		auth,
		mailConfig.EMail,
		[]string{to},
		[]byte(message),
	)

	if err != nil {
		panic(err)
	}
	return err
}

func parseConfig() {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	//config := Config{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("读取配置:%s\n", config)
}

//参考net/smtp的func SendMail()
//使用net.Dial连接tls(ssl)端口时,smtp.NewClient()会卡住且不提示err
//len(to)>1时,to[1]开始提示是密送
func SendMailUsingTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) (err error) {
	//create smtp client
	c, err := Dial(addr)
	if err != nil {
		log.Println("创建SMTP客户端错误:", err)
		return err
	}
	defer c.Close()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				log.Println("AUTH授权错误", err)
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

//return a smtp client
func Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		log.Println("Dialing Error:", err)
		return nil, err
	}
	//分解主机端口字符串
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

type Staff struct {
	IdCard string // 身份证号
	Name   string // 姓名
	Mail   string // 邮箱
	Mobile string // 手机号
}

type Config struct {
	Mail    MailConfig    `json:"mail"`
	Company CompanyConfig `json:"company"`
	Salary  SalaryConfig  `json:"salary"`
	Staff   StaffConfig   `json:"staff"`
}

type MailConfig struct {
	EMail    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Ssl      bool   `json:"ssl"`
}

type CompanyConfig struct {
	Name        string `json:"name"`
	SalaryPhone string `json:"salary-phone"`
}

type SalaryConfig struct {
	NoneHeader bool `json:"header"` // 默认false
}

type StaffConfig struct {
	NoneHeader bool `json:"header"` // 默认false
}

type SalaryTable struct {
	Header  string
	Content string
}

type Salary struct {
	Id                string // 人员编号
	Name              string // 姓名
	IdCard            string // 身份证号
	BasicWage         string // 基本工资
	PostWage          string // 岗位工资
	XBonus            string // 年资津贴
	WageHike          string // 加薪
	RedPackets        string // 开门红
	MeritPay          string // 绩效
	AllowanceSubsidy  string // 防寒暑费
	HolidaySubsidy    string // 节日费
	TotalPay          string // 应发合计
	BasicPension      string // 基本养老金
	MedicareBenefits  string // 医疗保险金
	InsuranceBenefits string // 失业保险金
	HousingFund       string // 住房公积金
	EnterpriseAnnuity string // 企业年金
	LaborFee          string // 工会费
	HouseRent         string // 房租费
	Electricity       string // 电费
	PropertyFee       string // 物业费
	SpecialDeduction  string // 专项扣除项目
	TaxBase           string // 扣税基数
	DeductOrRefundTax string // 补扣退个税
	CurrentTax        string // 本月扣税
	TotalDeduct       string // 实际扣款合计
	TotalIncome       string // 实发合计
	Signature         string // 签名
	BankName          string // 银行名称
	BankAccount       string // 银行账号
}
