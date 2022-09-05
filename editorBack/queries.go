package main

//queryAllPrograms get all programs
const queryAllPrograms string = `
{
	queryAllPrograms(func: has(Code)) {
		uid
		Code
	}
}`

// queryProgramByUid get one program by uid
const queryProgramByUid string = `
{
	node(func: uid(%s)) {
	  uid
	  Code
	  expand(_all_) {
		uid
		expand(_all_)
	  }
	}
}`

//queryAllPrograms save query to get all programs
const queryPaginationPrograms string = `
{
	queryPrograms(func: has(Code), first:2, after:%s) {
		uid
		Code
	}
}`
