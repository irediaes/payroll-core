package main

import (
	adm "github.com/ebikode/payroll-core/domain/admin"
	ast "github.com/ebikode/payroll-core/domain/app_setting"
	emp "github.com/ebikode/payroll-core/domain/employee"
	pyr "github.com/ebikode/payroll-core/domain/payroll"
	slr "github.com/ebikode/payroll-core/domain/salary"
	tax "github.com/ebikode/payroll-core/domain/tax"
	jb "github.com/ebikode/payroll-core/jobs"
	storage "github.com/ebikode/payroll-core/storage/mysql"
	"github.com/whiteshtef/clockwork"
)

// InitJobs Initialize all scheduled jobs
func InitJobs(mdb *storage.MDatabase) {

	var adminStorage adm.AdminRepository
	var employeeStorage emp.EmployeeRepository
	var payrollStorage pyr.PayrollRepository
	var taxStorage tax.TaxRepository
	var salaryStorage slr.SalaryRepository
	var appSettingStorage ast.AppSettingRepository

	// initalising all domain storage for db manipulation
	adminStorage = storage.NewDBAdminStorage(mdb)
	payrollStorage = storage.NewDBPayrollStorage(mdb)
	employeeStorage = storage.NewDBEmployeeStorage(mdb)
	salaryStorage = storage.NewDBSalaryStorage(mdb)
	taxStorage = storage.NewDBTaxStorage(mdb)
	appSettingStorage = storage.NewDBAppSettingStorage(mdb)

	// Initailizing application domain services
	admService := adm.NewService(adminStorage)
	empService := emp.NewService(employeeStorage)
	pyrService := pyr.NewService(payrollStorage)
	taxService := tax.NewService(taxStorage)
	salaryService := slr.NewService(salaryStorage)
	astService := ast.NewService(appSettingStorage)

	// Initialize clockwork schedules
	sched := clockwork.NewScheduler()

	// go runJobs(leaguesURL, fixtureURL, ls, cs, ss, fs, ts)
	var runJobs = func() {

		// RunCreateDefaultSuperAdmin - Create Default admin  if it doesn't exist
		// the first time the server is launch
		jb.RunCreateDefaultSuperAdmin(admService)

		// Create default app settings
		jb.RunCreateDefaultSettings(astService)

		// Create default employees
		jb.RunCreateDefaultEmployees(empService, salaryService)

		// Create Payroll Demo Data
		jb.RunDefaultPayrollGenerationJob(pyrService, empService, taxService)

		var runGeneratePayrollJob = func() {
			jb.RunPayrollGenerationJob(pyrService, astService, empService, taxService)
		}

		var runPayrollPaymentJob = func() {
			jb.RunPayrollPaymentJob(pyrService, astService)
		}

		runGeneratePayrollJob()
		runPayrollPaymentJob()

		// This runs every 12 Hours
		go sched.Schedule().Every(12).Hours().Do(runGeneratePayrollJob)

		// This runs every 60 Minutes
		go sched.Schedule().Every(60).Minutes().Do(runPayrollPaymentJob)

		sched.Run()
	}

	go runJobs()

}
