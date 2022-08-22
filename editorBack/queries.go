package main

//queryAllPrograms save query to get all programs
const queryAllPrograms string = `
{
	queryAllPrograms(func: has(Code)) {
		uid
		Code
	}
}`

// queryProgramByUid save query to get one program
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
