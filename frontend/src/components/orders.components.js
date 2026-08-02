import React, { useState, useEffect } from 'react';
import axios from "axios";
import { Button, Form, Container, Modal } from 'react-bootstrap';
import Order from './single-order.component';
import '../App.css';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'https://golangfullstackapp-a094n7l1.b4a.run';

const Orders = () => {
    const [orders, setOrders] = useState([]);
    const [errorMsg, setErrorMsg] = useState("");

    const [changeOrder, setChangeOrder] = useState({ change: false, id: 0 });
    const [changeWaiter, setChangeWaiter] = useState({ change: false, id: 0 });
    const [newWaiterName, setNewWaiterName] = useState("");

    const [addNewOrder, setAddNewOrder] = useState(false);
    const [newOrder, setNewOrder] = useState({ dish: "", server: "", table: "", price: 0, image: "" });
    const [selectedFile, setSelectedFile] = useState(null);

    useEffect(() => {
        getAllOrders();
    }, []);

    // uploads selected file to Go backend (/upload)
    async function uploadMediaFile() {
        if (!selectedFile) return newOrder.image || "";
        const formData = new FormData();
        formData.append("file", selectedFile);
        try {
            const res = await axios.post(`${API_BASE_URL}/upload`, formData, {
                headers: { 'Content-Type': 'multipart/form-data' }
            });
            return res.data && res.data.url ? res.data.url : "";
        } catch (err) {
            console.error("Error uploading file:", err);
            return "";
        }
    }

    // changes the waiter
    function changeWaiterForOrder() {
        const url = `${API_BASE_URL}/waiter/update/${changeWaiter.id}`;
        setChangeWaiter({ change: false, id: 0 });
        axios.put(url, {
            server: newWaiterName
        }).then(response => {
            if (response.status === 200) {
                getAllOrders();
            }
        }).catch(err => {
            console.error("Error updating waiter:", err);
            setErrorMsg("Gagal mengubah waiter. Pastikan server backend (Golang) berjalan.");
        });
    }

    // changes the order
    async function changeSingleOrder() {
        setChangeOrder({ change: false, id: 0 });
        try {
            const uploadedUrl = await uploadMediaFile();
            const url = `${API_BASE_URL}/order/update/${changeOrder.id}`;
            const orderToUpdate = {
                ...newOrder,
                price: parseFloat(newOrder.price) || 0,
                image: uploadedUrl || newOrder.image
            };
            axios.put(url, orderToUpdate)
                .then(response => {
                    if (response.status === 200) {
                        getAllOrders();
                        setNewOrder({ dish: "", server: "", table: "", price: 0, image: "" });
                        setSelectedFile(null);
                    }
                })
                .catch(err => {
                    console.error("Error updating order:", err);
                    setErrorMsg("Gagal mengubah order. Pastikan server backend (Golang) berjalan.");
                });
        } catch (err) {
            setErrorMsg("Gagal mengunggah file media ke server.");
        }
    }

    // creates a new order
    async function addSingleOrder() {
        setAddNewOrder(false);
        try {
            const uploadedUrl = await uploadMediaFile();
            const url = `${API_BASE_URL}/order/create`;
            axios.post(url, {
                server: newOrder.server,
                dish: newOrder.dish,
                table: newOrder.table,
                price: parseFloat(newOrder.price) || 0,
                image: uploadedUrl
            }).then(response => {
                if (response.status === 200) {
                    getAllOrders();
                    setNewOrder({ dish: "", server: "", table: "", price: 0, image: "" });
                    setSelectedFile(null);
                }
            }).catch(err => {
                console.error("Error adding order:", err);
                setErrorMsg("Gagal menambah order. Pastikan server backend (Golang) berjalan.");
            });
        } catch (err) {
            setErrorMsg("Gagal mengunggah file media ke server.");
        }
    }

    // gets all the orders
    function getAllOrders() {
        const url = `${API_BASE_URL}/orders`;
        axios.get(url, {
            responseType: 'json'
        }).then(response => {
            if (response.status === 200) {
                setOrders(response.data || []);
                setErrorMsg("");
            }
        }).catch(err => {
            console.error("Error fetching orders:", err);
            setErrorMsg(`Tidak dapat terhubung ke server (Network Error). Pastikan server Go API aktif di ${API_BASE_URL}.`);
        });
    }

    // deletes a single order
    function deleteSingleOrder(id) {
        const url = `${API_BASE_URL}/order/delete/${id}`;
        axios.delete(url)
            .then(response => {
                if (response.status === 200) {
                    getAllOrders();
                }
            })
            .catch(err => {
                console.error("Error deleting order:", err);
                setErrorMsg("Gagal menghapus order. Pastikan server backend (Golang) berjalan.");
            });
    }

    return (
        <div>
            {/* add new order button */}
            <Container className="mt-3">
                {errorMsg && (
                    <div className="alert alert-warning text-center" role="alert">
                        {errorMsg}
                    </div>
                )}
                <Button className='buttonAdd' onClick={() => { setSelectedFile(null); setAddNewOrder(true); }}>Add New Order</Button>
            </Container>

            {/* list all current orders */}
            <Container>
                {orders != null && orders.map((order, i) => (
                    <Order
                        key={order._id || i}
                        orderData={order}
                        deleteSingleOrder={deleteSingleOrder}
                        setChangeWaiter={setChangeWaiter}
                        setChangeOrder={setChangeOrder}
                    />
                ))}
            </Container>

            {/* popup for adding a new order */}
            <Modal className='customModal' show={addNewOrder} onHide={() => setAddNewOrder(false)} centered>
                <Modal.Header closeButton>
                    <Modal.Title className='modalTitle'>Add Order</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Form.Group>
                        <Form.Label className='customFormLabel'>Dish</Form.Label>
                        <Form.Control onChange={(event) => setNewOrder({ ...newOrder, dish: event.target.value })} />
                        <Form.Label className='customFormLabel'>Waiter</Form.Label>
                        <Form.Control onChange={(event) => setNewOrder({ ...newOrder, server: event.target.value })} />
                        <Form.Label className='customFormLabel'>Table</Form.Label>
                        <Form.Control onChange={(event) => setNewOrder({ ...newOrder, table: event.target.value })} />
                        <Form.Label className='customFormLabel'>Price</Form.Label>
                        <Form.Control type="number" onChange={(event) => setNewOrder({ ...newOrder, price: event.target.value })} />
                        <Form.Label className='customFormLabel mt-2'>Image / Media</Form.Label>
                        <Form.Control type="file" accept="image/*" onChange={(event) => setSelectedFile(event.target.files[0])} />
                    </Form.Group>
                </Modal.Body>
                <Modal.Footer>
                    <Button className='primary-button' onClick={() => addSingleOrder()}>Add</Button>
                    <Button className='secondary-button' onClick={() => setAddNewOrder(false)}>Cancel</Button>
                </Modal.Footer>
            </Modal>

            {/* popup for changing a waiter */}
            <Modal className='customModal' show={changeWaiter.change} onHide={() => setChangeWaiter({ change: false, id: 0 })} centered>
                <Modal.Header closeButton>
                    <Modal.Title className='modalTitle'>Change Waiter</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Form.Group>
                        <Form.Label className='customFormLabel'>New Waiter</Form.Label>
                        <Form.Control onChange={(event) => setNewWaiterName(event.target.value)} />
                    </Form.Group>
                </Modal.Body>
                <Modal.Footer>
                    <Button className='primary-button' onClick={() => changeWaiterForOrder()}>Change</Button>
                    <Button className='secondary-button' onClick={() => setChangeWaiter({ change: false, id: 0 })}>Cancel</Button>
                </Modal.Footer>
            </Modal>

            {/* popup for changing an order */}
            <Modal className='customModal' show={changeOrder.change} onHide={() => setChangeOrder({ change: false, id: 0 })} centered>
                <Modal.Header closeButton>
                    <Modal.Title className='modalTitle'>Change Order</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Form.Group>
                        <Form.Label className='customFormLabel'>Dish</Form.Label>
                        <Form.Control onChange={(event) => setNewOrder({ ...newOrder, dish: event.target.value })} />
                        <Form.Label className='customFormLabel'>Waiter</Form.Label>
                        <Form.Control onChange={(event) => setNewOrder({ ...newOrder, server: event.target.value })} />
                        <Form.Label className='customFormLabel'>Table</Form.Label>
                        <Form.Control onChange={(event) => setNewOrder({ ...newOrder, table: event.target.value })} />
                        <Form.Label className='customFormLabel'>Price</Form.Label>
                        <Form.Control type="number" onChange={(event) => setNewOrder({ ...newOrder, price: event.target.value })} />
                        <Form.Label className='customFormLabel mt-2'>Image / Media</Form.Label>
                        <Form.Control type="file" accept="image/*" onChange={(event) => setSelectedFile(event.target.files[0])} />
                    </Form.Group>
                </Modal.Body>
                <Modal.Footer>
                    <Button className='primary-button' onClick={() => changeSingleOrder()}>Change</Button>
                    <Button className='secondary-button' onClick={() => setChangeOrder({ change: false, id: 0 })}>Cancel</Button>
                </Modal.Footer>
            </Modal>
        </div>
    );
};

export default Orders;