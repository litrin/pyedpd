# pyedpd
A daemon for EMON files automation seeking and processing by pyedp tool

- edpd build up
  - Download and install the latest go lang package from https://golang.org/dl/.
  - Download source codes, go to source code's directory and run windows cmd : go build -ldflags="-s -w"
  
- Hardware requirments:
  - Windows host with MS office (intenal IT buid Win7/Win10 is suggested) or Linux.
  - Linux version only support to generate csv files.
  - At least 4G memory for each hyper-threads.
  - Ruby runtime enviroments and EDP are pre-installed in the host.  
  
- edpd setup
  - Copy edpd.exe to the EDP's directory on the edpd host. 
  - Make a new direcotry for Emon data files, for example: c:\emon-data.
  - Run windows command: edpd -emon c:\emon-data
  - To sharing the service, you may set the c:\emon-data as a windows share folder

- Emon data requiremts
  - 3 files must be containded by each of the emon data folder: emon.dat, emon-v.dat, emon-m.dat.
  - Following the limitations from excel, the file emon.dat had better less than 200M. 
  
- About EDPD_LOCK file 
  - When Emon data was on processing or process finished, is the same folder, there was a file named EDPD_LOCK.
  - Under windows, when Emon data processed, an other edpd_summary.xlsx file will exist in the Emon data's folder.
  - Archive feature enable system create a zip files for all emon raw data after processing finished.
