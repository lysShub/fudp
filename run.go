package fudp

import "errors"

func (f *Fudp) run() (err error) {

	if f.mode == PPMode {
		if f.role == CRole {
			return f.reveive()
		} else if f.role == SRole {
			return f.send()
		} else {
			return errors.New("invalid role")
		}

	} else if f.mode == CSMode {
		if f.role == CRole {
			return f.client()
		} else if f.role == SRole {

		} else {
			return errors.New("invalid role")
		}
	} else {
		return errors.New("invalid mode")
	}
	return
}

func (f *Fudp) send() (err error) {
	return
}

func (f *Fudp) reveive() (err error) { return }

func (f *Fudp) client() (err error) {
	return
}

func (f *Fudp) server() (err error) {
	return
}
