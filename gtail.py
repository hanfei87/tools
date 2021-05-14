
import os
import threading
import stat
import time
import re
import PySimpleGUI as psg

_layout = [
    [psg.FileBrowse('watch file', key='btn_add_file', size=(8, 1), disabled=False),
     psg.InputText('', key='txt_file_path')],

    [psg.Text('key word', key='txt_key_word'),
     psg.InputText('', key='txt_keyword', tooltip='key words to be matched, support regular expression.', size=(22, 1)),
     psg.Checkbox('show matched', key='chk_only_matched'),
     psg.Button('start', key='btn_start_stop', size=(8, 1), disabled=False),
     psg.Button('clear', key='btn_clear', size=(8, 1)), ],

    [psg.Multiline('', key='mline_output', size=(90, 40), autoscroll=True), ],
]


def main():
    timeout = 1000
    window = psg.Window('Windows tail', _layout)
    start_stop = False
    while True:
        event, value = window.Read(timeout=timeout)
        if event == 'btn_add_file':
            # this button do not make a event.
            pass
        elif event == 'btn_clear':
            window.Element('mline_output').update(value='')
        elif event == 'btn_start_stop':
            if value['btn_add_file'] != '':
                file_path = value['btn_add_file']
                if os.path.isfile(file_path) and \
                   stat.S_ISREG(os.stat(file_path).st_mode):

                    # start tailing
                    if start_stop is False:
                        window.Element('btn_start_stop').update('stop')
                        start_tail(file_path, window.Element('mline_output'))
                        start_stop = True
                    # stop tailing
                    else:
                        window.Element('btn_start_stop').update('start')
                        start_stop = False
                        stop_tail()
                else:
                    psg.PopupError(f'[{file_path}] is not regular file.')
            else:
                psg.Popup('select a target file to watch first.')

        elif event in (None, 'exit'):
            break

    window.Close()



_tail_thread_run = False
def stop_tail():
    global _tail_thread_run
    _tail_thread_run = False


def start_tail(file_path, output_elm):
    global _tail_thread_run
    _tail_thread_run = True
    t = threading.Thread(target=tail_thread, args=(file_path, output_elm))
    t.start()


def tail_thread(file, output_elm):
    global _tail_thread_run
    with open(file, 'r') as f:
        init_size = 0
        while _tail_thread_run is True:
            new_size = os.stat(file).st_size
            if new_size > init_size:
                try:
                    f.seek(init_size)
                    new_dat = f.read().replace('\n', '\n>')
                    init_size = new_size
                    output_elm.update(append=True, value=new_dat)
                except Exception as e:
                    psg.PopupError(str(e))
            else:
                time.sleep(0.2)


if __name__ == "__main__":
    main()
